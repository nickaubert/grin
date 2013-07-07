package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "code.google.com/p/goncurses"
import "fmt"

func main() {

	stdscr, _ := goncurses.Init()
	defer goncurses.End()

    well_depth := 20
    well_width := 10

    block_height    := well_depth
    block_longitude := ( well_width / 2 )

    well_border := define_well( well_width , well_depth )

    draw_well( well_border , stdscr )

    for keep_going := true ; keep_going == true ; {

        // erase_block( stdscr , well_border , block_height , block_longitude )
        draw_block( stdscr , well_border , "erase" , block_height , block_longitude )
        block_height--
        // operation := "draw"
        keep_going = draw_block( stdscr , well_border , "draw" , block_height , block_longitude )

        show_stats( stdscr , block_height )

        somechar := stdscr.GetChar()
        switch {
            case somechar == 113:
                keep_going = false
            // case somechar == 68:
            //     move_block( 'left' )
        }
        string_status := fmt.Sprintf( "string: %03d" , somechar )
        stdscr.MovePrint( 3 , 3  , string_status ) // TESTING

    }

	goncurses.End()

}

func define_well( well_width , well_depth int ) [][]byte {
    well := make([][]byte, well_depth)
    for i := 0 ; i < well_depth ; i++ {
        well_row := make([]byte, well_width)
        well[i] = well_row
    }
    return well
}

func show_stats( stdscr goncurses.Window , block_height int ) {

    bh_status := fmt.Sprintf( "block height: %02d" , block_height )
    stdscr.MovePrint( 1 , 1  , bh_status )

}

func erase_block( stdscr goncurses.Window , this_well [][]byte , block_height , block_longitude int ) {

    // terminal size
    // term_row, term_col := stdscr.Maxyx()
    _, term_col := stdscr.Maxyx()

    well_top    := 5
    well_height := len(this_well)
    well_width  := len(this_well[1])
    well_left   := ( ( term_col / 2 ) - well_width )

    stdscr.MovePrint( ( well_top + well_height - block_height ) ,  ( well_left + block_longitude )  , " " )

}

func draw_block( stdscr goncurses.Window , this_well [][]byte , operation string , block_height , block_longitude int ) bool {

    // terminal size
    // term_row, term_col := stdscr.Maxyx()
    _, term_col := stdscr.Maxyx()

    well_top    := 5
    well_height := len(this_well)
    well_width  := len(this_well[1])
    well_left   := ( ( term_col / 2 ) - well_width )

    block_paint := "XX"
    if operation == "erase" {
        block_paint = "  "
    }
    stdscr.MovePrint( ( well_top + well_height - block_height ) , ( well_left + block_longitude )  , block_paint )

    if block_height == 0 {
        return false
    }
    return true

}


func draw_well( this_well [][]byte , stdscr goncurses.Window  ) {

    // terminal size
    // term_row, term_col := stdscr.Maxyx()
    _, term_col := stdscr.Maxyx()

    // sides
    well_height := len(this_well)
    well_width  := len(this_well[1])
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

