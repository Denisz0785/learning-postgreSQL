package main

import (
	"context"
	"expenses/repository"
	"flag"
	"fmt"
	"strings"
)

func main() {

	// ConnectToDB connects to DB
	myUrl := "MYURL"
	ctx := context.Background()
	conn, err := repository.ConnectToDB(ctx, myUrl)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer conn.Close(ctx)

	// Create new structure wich consist data about connection with database
	ConnExpRepo := repository.NewExpenseRepo(conn)

	// define flags for getting values of flags command ./expenses cmd=get_expense_types user=Ivan and
	// ./expenses -cmd=Add -login=igor23 -exp_type=swimming -time=2024-02-25-17:26 -spent=500
	funcPtr := flag.String("cmd", "none", "function")
	userPtr := flag.String("name", "none", "user's name")
	loginPtr := flag.String("login", "none", "user's login")
	expTypePtr := flag.String("exp_type", "none", "type of expenses")
	timePtr := flag.String("time", "none", "time of expenses")
	spentPtr := flag.Float64("spent", 0.0, "amount of expenses")
	//Parse() parses the command line into the defined flags
	flag.Parse()

	// define which command was input
	switch {
	case strings.EqualFold(*funcPtr, "Get_ManyRows"):
		var resultExpenses []string
		resultExpenses, err = ConnExpRepo.GetUserExpenseTypes(*userPtr)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println()
		fmt.Printf("Expenses_type of %s = %s\n", *userPtr, resultExpenses)

	case strings.EqualFold(*funcPtr, "add"):
		//
		err := ConnExpRepo.AddExpTransaction(ctx, loginPtr, expTypePtr, timePtr, spentPtr)
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("check your input data in a command-line")
	}

}
