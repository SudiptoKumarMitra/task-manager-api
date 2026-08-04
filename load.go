package main
import(
"os"
"encoding/json"
)
func loadTasks() error {
	file,err := os.Open("tasklist.json")
	if err != nil {
		if os.IsNotExist(err) {
			Tasks = []Task{}
			return nil
		}
		return err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&Tasks)
	if err != nil {
		return err
	}
	for i := range Tasks {
		if Tasks[i].ID >= nextID {
			nextID = Tasks[i].ID + 1
		}
	}
	return err
}