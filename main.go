package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sync"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})
	api := NewItemAPIService(NewMemoryItemRespository([]Item{
		{Id: "1"},
		{Id: "2"},
		{Id: "3"},
	}))
	mux.HandleFunc("GET /items", api.getAllItems)
	mux.HandleFunc("GET /items/{id}", api.getItemById)
	mux.HandleFunc("DELETE /items/{id}", api.deleteItemById)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

type ItemRepository interface {
	getAllItems() []Item
	getItemById(id string) (Item, bool)
	insertItem(item Item) error
	deleteItemById(id string) (Item, error)
}

type MemoryItemRepository struct {
	items []Item
	mu    sync.Mutex
}

func NewMemoryItemRespository(items []Item) *MemoryItemRepository {
	return &MemoryItemRepository{items: items}
}

func (ir *MemoryItemRepository) getAllItems() []Item {
	defer ir.mu.Unlock()
	ir.mu.Lock()
	return ir.items
}

func (ir *MemoryItemRepository) getItemById(id string) (Item, bool) {
	defer ir.mu.Unlock()
	ir.mu.Lock()
	for i := range ir.items {
		item := ir.items[i]
		if item.Id == id {
			return item, true
		}
	}
	return Item{}, false
}

func (ir *MemoryItemRepository) insertItem(item Item) error {
	defer ir.mu.Unlock()
	ir.mu.Lock()
	ir.items = append(ir.items, item)
	return nil
}

func (ir *MemoryItemRepository) deleteItemById(id string) (Item, error) {
	defer ir.mu.Unlock()
	ir.mu.Lock()
	index := slices.IndexFunc(ir.items, func(currentItem Item) bool {
		return id == currentItem.Id
	})
	if index < 0 {
		return Item{}, nil
	}
	deletedItem := ir.items[index]
	ir.items = slices.Delete(ir.items, index, min(index+1, len(ir.items)))
	return deletedItem, nil
}

type Item struct {
	Id string `json:"id"`
}

type ItemAPIService struct {
	repository ItemRepository
}

func NewItemAPIService(repository ItemRepository) *ItemAPIService {
	return &ItemAPIService{repository: repository}
}

func (api *ItemAPIService) getAllItems(w http.ResponseWriter, r *http.Request) {
	items := api.repository.getAllItems()
	writeJSON(w, items)
}

func (api *ItemAPIService) getItemById(w http.ResponseWriter, r *http.Request) {
	requestedId := r.PathValue("id")
	item, ok := api.repository.getItemById(requestedId)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, item)
}

func (api *ItemAPIService) deleteItemById(w http.ResponseWriter, r *http.Request) {
	idToDelete := r.PathValue("id")
	if len(idToDelete) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deltedItem, err := api.repository.deleteItemById(idToDelete)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, deltedItem)
}

func writeJSON(w http.ResponseWriter, object any) {
	jsonObj, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonObj)
}
