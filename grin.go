package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "code.google.com/p/goncurses"
import "fmt"

func main() {

	stdscr, _ := goncurses.Init()
	defer goncurses.End()

    term_row, term_col := stdscr.Maxyx()
	goncurses.End()

    well_width := 10
    well_depth := 20

    well_border := define_well( well_width , well_depth )

    draw_well( well_border , term_row , term_col)

    // somechar := stdscr.GetChar()
	// goncurses.End()

}

func define_well( well_width , well_depth int ) [][]byte {
    well := make([][]byte, well_depth)
    for i := 0 ; i < well_depth ; i++ {
        well_row := make([]byte, well_width)
        well[i] = well_row
    }
    return well
}

func draw_well( this_well [][]byte , term_row , term_col int ) {

    // offset from left
    var line_buffer string
    columns_offset := ( ( term_col / 2 ) - len(this_well[1]) - 1 )
    for i := 0 ; i < columns_offset ; i++ {
        line_buffer += " "
    }

    // draw well sides
    var well_width int
    for _, this_row := range this_well {
        this_line := line_buffer
        this_line += "║"
        for _, c := range this_row {
            if c == 0 {
                this_line += "  "
            } else {
                this_line += "++" // brick?
            }
        }

        this_line += "║"
        fmt.Println( this_line )

        well_width = len( this_row )

    }

    // draw well bottom
    this_line := line_buffer
    this_line += "╩"
    for i := 0 ; i < well_width ; i++ {
        this_line += "══"
    }
    this_line += "╩"
    fmt.Println( this_line )

    fmt.Println( "hello" )
    fmt.Printf( "one" )
    fmt.Printf( "two" )

}

