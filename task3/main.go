package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func filterUsers(users []User, filter func(user User) bool) (result []User) {
	for _, user := range users {
		if filter(user) {
			result = append(result, user)
		}
	}
	return
}

func readCollection(fileName string) (result []User, err error) {
	text, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			//suppress this warning, if file is not presented then collection is empty
			return result, nil
		}
		return
	}
	err = json.Unmarshal(text, &result)
	return
}

func write(users []User, fileName string) (err error) {
	content, err := json.Marshal(users)
	if err != nil {
		return
	}
	return ioutil.WriteFile(fileName, content, 777)
}

var operations = map[string]func(Arguments, io.Writer) error{
	"list": func(arguments Arguments, writer io.Writer) (err error) {
		data, err := ioutil.ReadFile(arguments["fileName"])
		if err != nil {
			return
		}
		_, err = writer.Write(data)
		return
	},
	"remove": func(arguments Arguments, writer io.Writer) (err error) {
		if len(arguments["id"]) == 0 {
			return errors.New("-id flag has to be specified")
		}

		users, err := readCollection(arguments["fileName"])
		if err != nil {
			return
		}

		result := filterUsers(users, func(user User) bool {
			return user.Id != arguments["id"]
		})

		if len(users) == len(result) {
			_, err = writer.Write([]byte("Item with id " + arguments["id"] + " not found"))
			return
		}

		return write(result, arguments["fileName"])
	},
	"add": func(arguments Arguments, writer io.Writer) (err error) {
		if len(arguments["item"]) == 0 {
			return errors.New("-item flag has to be specified")
		}

		var userToAdd User
		if err = json.Unmarshal([]byte(arguments["item"]), &userToAdd); err != nil {
			return
		}

		users, err := readCollection(arguments["fileName"])
		if err != nil {
			return
		}

		var usersWithSamePK = filterUsers(users, func(user User) bool {
			return user.Id == userToAdd.Id
		})

		if len(usersWithSamePK) != 0 {
			_, err = writer.Write([]byte("Item with id " + userToAdd.Id + " already exists"))
			return
		}

		return write(append(users, userToAdd), arguments["fileName"])
	},
	"findById": func(arguments Arguments, writer io.Writer) (err error) {
		if len(arguments["id"]) == 0 {
			return errors.New("-id flag has to be specified")
		}

		users, err := readCollection(arguments["fileName"])
		if err != nil {
			return
		}

		var filteredUsers = filterUsers(users, func(user User) bool {
			return user.Id == arguments["id"]
		})

		if len(filteredUsers) == 0 {
			_, err = writer.Write([]byte(""))
			return
		}

		content, err := json.Marshal(filteredUsers[0])
		if err != nil {
			return
		}

		_, err = writer.Write(content)
		return

	},
}

func Perform(arguments Arguments, writer io.Writer) (err error) {
	if len(arguments["operation"]) == 0 {
		return errors.New("-operation flag has to be specified")
	}
	if len(arguments["fileName"]) == 0 {
		return errors.New("-fileName flag has to be specified")
	}

	operation, operationExists := operations[arguments["operation"]]
	if !operationExists {
		return errors.New("Operation " + arguments["operation"] + " not allowed!")
	}

	return operation(arguments, writer)
}

func parseArgs() Arguments {
	operation := flag.String("operation", "", "Specify operation")
	id := flag.String("id", "", "Specify id")
	item := flag.String("item", "", "Specify item")
	fileName := flag.String("fileName", "", "Specify fileName")
	flag.Parse()
	return Arguments{"operation": *operation, "id": *id, "item": *item, "fileName": *fileName}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
