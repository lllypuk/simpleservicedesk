package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
    ID   string `json:"_id,omitempty"`
    Text string `json:"text"`
}

var client *mongo.Client

func init() {
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    var err error
    client, err = mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        panic(err)
    }
    fmt.Println("Connected to MongoDB!")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("HX-Request") == "true" {
        fmt.Fprint(w, "This is an htmx response!")
        return
    }
    fmt.Fprintln(w, "Hello, World!")
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
    collection := client.Database("todoapp").Collection("todos")
    var todos []Todo
    cursor, err := collection.Find(context.Background(), bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if err = cursor.All(context.Background(), &todos); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(todos)
}

func addTodoHandler(w http.ResponseWriter, r *http.Request) {
    var todo Todo
    if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    collection := client.Database("todoapp").Collection("todos")
    result, err := collection.InsertOne(context.Background(), todo)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    todo.ID = result.InsertedID.(primitive.ObjectID).Hex()
    json.NewEncoder(w).Encode(todo)
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/", helloHandler)
    r.HandleFunc("/todos", todosHandler).Methods("GET")
    r.HandleFunc("/add", addTodoHandler).Methods("POST")

    fmt.Println("Server listening on :8081")
    if err := http.ListenAndServe(":8081", r); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
