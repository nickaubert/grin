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

        show_stats( stdscr , block_height )

        // keyboard input
        somechar := stdscr.GetChar()
        movement := "hold"
        switch {
            case somechar == 113:
                keep_going = false
            case somechar == 106 :
                movement = "left"
            case somechar == 108 :
                movement = "right"
            case somechar == 32 :
                movement = "drop"
        }
        string_status := fmt.Sprintf( "string: %03d" , somechar )
        stdscr.MovePrint( 3 , 3  , string_status ) // TESTING

        // erase old block
        draw_block( stdscr , well_border , "erase" , block_height , block_longitude )

        // move block 
        block_height--
        block_height , block_longitude = move_block( well_border , movement  , block_height , block_longitude )

        // draw new block
        keep_going = draw_block( stdscr , well_border , "draw" , block_height , block_longitude )

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

func move_block( this_well [][]byte , operation string , block_height , block_longitude int ) ( int , int ) {

    well_width  := len(this_well[1])

    switch {
        case operation == "left" :
            if block_longitude > 1 {
                block_longitude--
            }
        case operation == "right" :
            if block_longitude < ( well_width - 1 ) {
                block_longitude++
            }
        case operation == "drop" :
            block_height = 0 // need collision detection here
    }
    return block_height , block_longitude
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
    stdscr.MovePrint( ( well_top + well_height - block_height ) , ( well_left + ( block_longitude * 2 ) )  , block_paint )

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
        stdscr.MovePrint( row_height , well_left  , " |" )
        stdscr.MovePrint( row_height , well_right ,  "|" )
    }

    for col_loc := well_left ; col_loc <= well_right ; col_loc ++ {
        stdscr.MovePrint( well_bottom , col_loc  , "=" )
    }

    stdscr.Refresh()

}

