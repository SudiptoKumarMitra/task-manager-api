package main
import (
	"encoding/json"
	"net/http"
)

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("id") != "" {
		id, err := getID(r)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		var task Task
		err = DB.QueryRow("SELECT id,title FROM tasks WHERE id = $1",id).Scan(&task.ID,&task.Title)
		if err != nil {
			http.Error(w, "Task Not Found", http.StatusNotFound)
			return
		}
		headerSet(w)
		json.NewEncoder(w).Encode(task)
	} else {
		rows,err := DB.Query("SELECT id,title FROM tasks")
		if err != nil {
			http.Error(w, "Failed to connect to Database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var tasks []Task
		for rows.Next() {
			var task Task
			err := rows.Scan(&task.ID, &task.Title)
				if err != nil {
				http.Error(w, "Error in fetching data", http.StatusInternalServerError)
				return
			}
			tasks=append(tasks,task)
		}
		headerSet(w)
		json.NewEncoder(w).Encode(tasks)
	}
}

func handlePostTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil{
		http.Error(w, "Invalid JSON", http.StatusInternalServerError)
		return
	}

	var id int
	err = DB.QueryRow("INSERT INTO tasks (title) VALUES ($1) RETURNING id",task.Title).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to Insert into database",http.StatusInternalServerError)
		return
	}
	task.ID = id
	headerSet(w)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
	if err != nil{
		http.Error(w, "Invalid JSON", http.StatusInternalServerError)
		return
	}
}

func handlePutTask(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("id") != "" {
		id, err := getID(r)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		var task Task
		err = json.NewDecoder(r.Body).Decode(&task)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
		res,err := DB.Exec("UPDATE tasks SET title = $1 WHERE id = $2",task.Title,id)
		if err != nil {
			http.Error(w, "Error in updating data",http.StatusInternalServerError)
			return
		}
		affected,err := res.RowsAffected()
		if err != nil {
			http.Error(w,"Error in database",http.StatusInternalServerError)
			return
		}
		if affected == 0 {
			http.Error(w,"NO task exists with this ID",http.StatusNotFound)
			return
		}
		task.ID=id
		headerSet(w)
		err = json.NewEncoder(w).Encode(task)
		if err != nil {
			http.Error(w,"Error in encoding JSON",http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Task Not Found", http.StatusNotFound)
	}
}
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("id") != "" {
			id, err := getID(r)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			res,err := DB.Exec("DELETE FROM tasks WHERE id = $1",id)
			if err != nil {
				http.Error(w,"Error Occured in database",http.StatusInternalServerError)
				return
			}
			affected,err := res.RowsAffected()
			if err != nil {
				http.Error(w,"Error Occured in database", http.StatusInternalServerError)
				return
			}
			if affected == 0 {
				http.Error(w,"Task NOt Found", http.StatusNotFound)
				return
			}
			headerSet(w)
			err = json.NewEncoder(w).Encode("message :Task Deleted")
			if err != nil {
				http.Error(w,"Error Occured in encoding JSON",http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w,"Task NOt Found", http.StatusBadRequest)
		}

}