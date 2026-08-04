package main
import(
	"os"
	"encoding/json"
)
func saveTasks() error {
	jsondata, err := json.MarshalIndent(Tasks, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile("tasklist.json", jsondata, 0644)
	return err
}