package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "code.google.com/p/goncurses"
import "fmt"

func main() {

	stdscr, _ := goncurses.Init()
	defer goncurses.End()

    term_row, term_col := stdscr.Maxyx()
	// stdscr.Print("Hello, World %d %d !!!" , term_row , term_col )
	goncurses.End()

    well_width := 10
    well_depth := 20

    well_border := define_well( well_width , well_depth )

    draw_well( well_border , term_row , term_col)

}

func define_well( well_width , well_depth int ) []byte {
    // fmt.Println( well_width )
    // fmt.Println( well_depth )
    well := make([]byte, well_width)
    return well
}

func draw_well( this_well []byte , term_row , term_col int ) {
    // fmt.Println( term_row )
    // fmt.Println( term_col )

    var this_line string

    columns_offset := ( ( term_col / 2 ) - ( len(this_well) / 2 ) - 1 )
    for i := 0 ; i < columns_offset ; i++ { 
        this_line += " "
    } 

    // var this_line string
    this_line += "║"
    for _, c := range this_well {
        if c == 0 {
            this_line += " "
        } else {
            this_line += "+"
        }
    }
    this_line += "║"
    fmt.Println( this_line )
}

