package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

var jobs chan Drone

type Drone struct {
	ID      int64
	Size    int64
	Motor   int64
	Blade   int64
	Battery int64
}

func DroneDB(drone chan Drone) {
	fmt.Println("try bolt!!!!")
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo("bolt://localhost:7687")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("bolt!!!!")
	//var resultdrone Drone
	time.Sleep(time.Second)
	i := 1

	for {
		j := <-drone
		fmt.Print(j)
		i++
		//time.Sleep(3 * time.Second)
		// Here we prepare a new statement. This gives us the flexibility to
		// cancel that statement without any request sent to Neo
		stmt, err := conn.PrepareNeo("CREATE (n:NODE {ID: {id}})")
		if err != nil {
			panic(err)
		}

		// Executing a statement just returns summary information
		result, err := stmt.ExecNeo(map[string]interface{}{"id": j.ID})

		//result, err := stmt.ExecNeo(map[string]interface{}{"ID": drone.ID, "Size": drone.Size, "Motor": drone.Motor, "Blade": drone.Blade, "Battery": drone.Battery})
		if err != nil {
			panic(err)
		}

		numResult, err := result.RowsAffected()
		if err != nil {
			panic(err)
		}
		fmt.Printf("CREATED ROWS: %d\n", numResult) // CREATED ROWS: 1

		// Closing the statment will also close the rows
		stmt.Close()
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func DroneInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Drone Information is not correct"+"\n"+r.Method+" not allowed", http.StatusMethodNotAllowed)
		return
	}

	var drone Drone
	if err := json.NewDecoder(r.Body).Decode(&drone); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var resultdrone Drone
	resultdrone.ID = drone.ID
	resultdrone.Size = drone.Size
	resultdrone.Motor = drone.Motor
	resultdrone.Blade = drone.Blade
	resultdrone.Battery = drone.Battery

	//go DroneDB(resultdrone)
	jobs <- resultdrone
	fmt.Print(resultdrone)
	if err := json.NewEncoder(w).Encode(resultdrone); err != nil {
		log.Println(err)
		http.Error(w, "oops", http.StatusInternalServerError)
	}
}

func main() {

	jobs = make(chan Drone, 1)
	go DroneDB(jobs)

	r := mux.NewRouter()

	// Routes consist of a path and a handler function.
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/home", homeHandler)
	r.HandleFunc("/drone", DroneInfoHandler)
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":80", r))

}
