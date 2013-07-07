package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "code.google.com/p/goncurses"
import "fmt"

func main() {

	stdscr, _ := goncurses.Init()
	defer goncurses.End()

	// goncurses.End()

    well_width := 10
    well_depth := 20

    well_border := define_well( well_width , well_depth )

    draw_well( well_border , stdscr )

    somechar := stdscr.GetChar()
	goncurses.End()

    fmt.Println( "some char" )
    fmt.Println( somechar )

}

func define_well( well_width , well_depth int ) [][]byte {
    well := make([][]byte, well_depth)
    for i := 0 ; i < well_depth ; i++ {
        well_row := make([]byte, well_width)
        well[i] = well_row
    }
    return well
}

func draw_well( this_well [][]byte , stdscr goncurses.Window  ) {

    /*
    // offset from left
    var line_buffer string
    columns_offset := ( ( term_col / 2 ) - len(this_well[1]) - 1 )
    for i := 0 ; i < columns_offset ; i++ {
        line_buffer += " "
    }
    */

    // terminal size
    // term_row, term_col := stdscr.Maxyx()
    _, term_col := stdscr.Maxyx()

    // sides
    well_width  := len(this_well[1])
    well_height := len(this_well)
    well_left   := ( ( term_col / 2 ) - well_width )
    well_right  := ( well_left + ( well_width * 2 ) )
    well_top    := 5
    well_bottom := ( well_top + well_height )

    // draw sides
    for row_height := well_top ; row_height < well_bottom ; row_height ++ {
        stdscr.MovePrint( row_height , well_left  , "|" )
        stdscr.MovePrint( row_height , well_right , "|" )
    }

    for col_loc := well_left ; col_loc <= well_right ; col_loc ++ {
        stdscr.MovePrint( well_bottom , col_loc  , "=" )
    }

    stdscr.Refresh()

}

