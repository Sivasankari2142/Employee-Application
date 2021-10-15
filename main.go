package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"time"

	"math/rand"

	"log"

	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)



var client *mongo.Client

func db() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:37017")// Connect to //MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}

var userCollection = db().Database("userdb").Collection("userinfo")
type User struct {
	ID 			int		`json:"id,omitempty" bson:"id,omitempty"`
	Firstname      string `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname       string `json:"lastname,omitempty" bson:"lastname,omitempty"`
	Salary         int    `json:"salary,omitempty" bson:"salary,omitempty"`
}

func createProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

		rand.Seed(time.Now().UnixNano())
		min := 1
		max := 1000
		id := rand.Intn(max-min) + min
		var person User
		person = User{ID: id}
		err := json.NewDecoder(r.Body).Decode(&person)
		if err != nil {
			fmt.Print(err)
		}
		insertResult, err := userCollection.InsertOne(context.TODO(), person)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(insertResult.InsertedID)
	}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("userdb").Collection("userinfo")
	w.Header().Set("Content-Type", "application/json")

	type updateBody struct {
		ID json.Number `json:"userid"`
		Firstname string `json:firstname`
		Lastname string `json:lastname`
		Salary json.Number `json:"salary"`
	}
	var body updateBody
	e := json.NewDecoder(r.Body).Decode(&body)
	if e != nil {
		fmt.Print(e)
	}
	filter := bson.D{
		{"id", body.ID}}
	after := options.After
	returnOpt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	update := bson.D{{"$set", bson.D{{"firstname", body.Firstname},{"lastname",body.Lastname},{"salary",body.Salary}}}}
	updateResult := collection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
	var result primitive.M
	_ = updateResult.Decode(&result)
	json.NewEncoder(w).Encode(result)
}
func deleteProfile(response http.ResponseWriter, request *http.Request) {

		response.Header().Set("content-type", "application/json")
		params := mux.Vars(request)
		collection := client.Database("userdb").Collection("userinfo")
		id, _ := strconv.Atoi(params["id"])
		var user User
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := collection.FindOneAndDelete(ctx,User{ID : id}).Decode(&user)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{"message": "` + err.Error() + `" }`))
			return
		}
		json.NewEncoder(response).Encode(user)
}

func GetPersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	collection := client.Database("userdb").Collection("userinfo")
	id, _ := strconv.Atoi(params["id"])
	var user User
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx,User{ID : id}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(user)
	}


func getAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var results []primitive.M                                   //slice for multiple documents
	cur, err := userCollection.Find(context.TODO(), bson.D{{}}) //returns a *mongo.Cursor
	if err != nil {

		fmt.Println(err)

	}
	for cur.Next(context.TODO()) { //Next() gets the next document for corresponding cursor

		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem) // appending document pointed by Next()
	}
	cur.Close(context.TODO()) // close the cursor once stream of documents has exhausted
	json.NewEncoder(w).Encode(results)
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:37017")
	client, _ = mongo.Connect(ctx, clientOptions)
	s := mux.NewRouter()
	s.HandleFunc("/user", createProfile).Methods("POST")
	s.HandleFunc("/updateProfile", updateProfile).Methods("PUT")
	s.HandleFunc("/del/{id}", deleteProfile).Methods("GET")
	s.HandleFunc("/oneuser/{id}", GetPersonEndpoint).Methods("GET")
	s.HandleFunc("/users", getAllUsers).Methods("GET")
	s.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))
	log.Fatal(http.ListenAndServe(":8888", s) )// Run Server
}
