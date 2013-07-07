package main

import "code.google.com/p/goncurses"
import "fmt"

func main() {
	stdscr, _ := goncurses.Init()
	defer goncurses.End()

    row, col := stdscr.Maxyx()
	stdscr.Print("Hello, World %d %d !!!" , row , col )

    // msg := "Just a string "
    // stdscr.MovePrint(row/2, (col-len(msg))/2, msg)

    // stdscr.MovePrint(row-3, 0, "This screen has %d rows and %d columns. ", row, col)
    // stdscr.MovePrint(row-2, 0, "Try resizing your terminal window and then "+
    //             "run this program again.")
    // stdscr.Refresh()
    // stdscr.GetChar()

	// stdscr.Print("Hello, World %d %d !!!" , row , col )
	// stdscr.Refresh()
	// somechar := stdscr.GetChar()

	// stdscr.Refresh()
    // stdscr.AddChar(goncurses.Character(somechar))

    // if goncurses.Character(somechar) == 'z' {
    //     stdscr.Print("The Z key was pressed.")
    // }
	// stdscr.GetChar()

	goncurses.End()

    // this_draw := draw_well
    // fmt.Println( this_draw )

    fmt.Println( "this num is ║" )

    this_num := draw_well()
    fmt.Println( this_num )

    that_draw := 13
    fmt.Println( that_draw )

    fmt.Println( 55 )
}

func draw_well() (s int) {
    well := make([]byte, 5)
    for _, c := range well {
        fmt.Println(c)
    }
    // fmt.Println( "hohoho" )
    fmt.Println("whatsup")
    // fmt.Println( s )
    s = 12
    return s
}

