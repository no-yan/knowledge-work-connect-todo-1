package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/google/uuid"
	"github.com/no-yan/knowledge-work-connect-todo/gen/todo/v1/todov1connect" // generated by protoc-gen-connect-go

	todov1 "github.com/no-yan/knowledge-work-connect-todo/gen/todo/v1" // generated by protoc-gen-go
)

var m = sync.Map{}

type TodoServer struct{}

type Todo struct {
	Id     string
	Title  string
	Status bool
}

func (s *TodoServer) Add(
	ctx context.Context,
	req *connect.Request[todov1.AddRequest],
) (*connect.Response[todov1.AddResponse], error) {
	log.Println("Request headers: ", req.Header())

	id := uuid.New().String()
	m.Store(id, Todo{
		Id:     id,
		Title:  req.Msg.Title,
		Status: false,
	})

	res := connect.NewResponse(&todov1.AddResponse{
		Id:     id,
		Status: false,
	})
	res.Header().Set("Todo-Version", "v1")
	return res, nil
}

func (s *TodoServer) Delete(
	ctx context.Context,
	req *connect.Request[todov1.DeleteRequest],
) (*connect.Response[todov1.DeleteResponse], error) {
	log.Println("Request headers: ", req.Header())

	item, ok := m.Load(req.Msg.Id)
	if !ok {
		fmt.Printf("Todo (id %s) is not found", req.Msg.Id)
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo (id %s) not found", req.Msg.Id))
	}
	res := connect.NewResponse(&todov1.DeleteResponse{
		Id: item.(Todo).Id,
	})
	res.Header().Set("Todo-Version", "v1")
	return res, nil
}

func (s *TodoServer) Update(
	ctx context.Context,
	req *connect.Request[todov1.UpdateRequest],
) (*connect.Response[todov1.UpdateResponse], error) {
	log.Println("Request headers: ", req.Header())

	item, ok := m.Load(req.Msg.Id)
	if !ok {
		fmt.Printf("Todo (id %s) is not found", req.Msg.Id)
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo (id %s) not found", req.Msg.Id))
	}

	res := connect.NewResponse(&todov1.UpdateResponse{
		Id:     item.(Todo).Id,
		Status: false,
	})
	res.Header().Set("Todo-Version", "v1")
	return res, nil
}

func main() {
	todoer := &TodoServer{}
	mux := http.NewServeMux()

	path, handler := todov1connect.NewTodoServiceHandler(todoer)
	mux.Handle(path, handler)
	log.Fatal(http.ListenAndServe(
		"localhost:8080",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	))
}
