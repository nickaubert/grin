package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "code.google.com/p/goncurses"
import "fmt"

func main() {

    // curses
	stdscr, _ := goncurses.Init()
	defer goncurses.End()

    // define well
    well_dimensions := make( []int , 3 )
    well_dimensions[0] = 20 // well_depth
    well_dimensions[1] = 10 // well_width
    well_dimensions[2] = 5  // vert_headroom

    // starting block location
    block_location  := make( []int , 2 )
    block_location[0] = well_dimensions[0]     // block_height
    block_location[1] = well_dimensions[1] / 2 // block_longitude

    draw_border( stdscr , well_dimensions )

    for keep_going := true ; keep_going == true ; {

        show_stats( stdscr , block_location[0] )

        // keyboard input
        //  wait to drop time here?
        somechar := stdscr.GetChar()
        string_status := fmt.Sprintf( "string: %03d" , somechar )
        stdscr.MovePrint( 3 , 3  , string_status ) // TESTING

        // process input
        movement := "hold"
        switch {
            case somechar == 113 : // q
                keep_going = false
            case somechar == 106 : // j
                movement = "left"
            case somechar == 108 : // l
                movement = "right"
            case somechar == 110 : // n
                movement = "dropone"
            case somechar == 32 :  // [space]
                movement = "drop"
        }

        if keep_going == false {
            break
        }

        // move block 
        block_status := move_block( stdscr , well_dimensions , movement , block_location )

        // new block?
        if block_status == 2 {
            new_block( well_dimensions , block_location )
        }


    }

	goncurses.End()

}


func show_stats( stdscr goncurses.Window , block_height int ) {

    bh_status := fmt.Sprintf( "block height: %02d" , block_height )
    stdscr.MovePrint( 1 , 1  , bh_status )

}

func move_block( stdscr goncurses.Window , well_dimensions [] int , operation string , block_location []int ) int {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    blocked := check_collisions( well_dimensions , block_location , operation )

    if blocked == true {
        if operation == "dropone" {
            return 2
        } else {
            return 1
        }
    }

    draw_block( stdscr , well_dimensions , "erase" , block_location )

    switch {
        case operation == "left" :
            block_longitude--
        case operation == "right" :
            block_longitude++
        case operation == "dropone" :
            block_height--
        case operation == "drop" :
            block_height = 1
    }

    block_location[0] = block_height
    block_location[1] = block_longitude

    draw_block( stdscr , well_dimensions , "draw" , block_location )

    return 0

}

func check_collisions( well_dimensions , block_location []int , operation string ) bool {

    blocked := false
    switch {
        case operation == "left" :
            if block_location[1] == 1 {
                blocked = true
            }
        case operation == "right" :
            if block_location[1] == ( well_dimensions[1] - 1 ) {
                blocked = true
            }
        case operation == "dropone" :
            if block_location[0] == 1 {
                blocked = true
            }
        case operation == "drop" :
            blocked = false // nothing to do here yet
    }

    return blocked
}

func draw_block( stdscr goncurses.Window , well_dimensions []int , operation string , block_location []int ) bool {

    // terminal size
    // term_row, term_col := stdscr.Maxyx()
    _, term_col := stdscr.Maxyx()

    // well_top    := 5
    well_left   := ( ( term_col / 2 ) - well_dimensions[1] )

    block_height    := block_location[0]
    block_longitude := block_location[1]

    block_paint := "XX"
    if operation == "erase" {
        block_paint = "  "
    }
    stdscr.MovePrint( ( well_dimensions[2] + well_dimensions[0] - block_height ) , ( well_left + ( block_longitude * 2 ) )  , block_paint )

    if block_height == 0 {
        return false
    }
    return true

}

func draw_border( stdscr goncurses.Window , well_dimensions []int ) {

    // terminal size
    // term_row, term_col := stdscr.Maxyx()
    _, term_col := stdscr.Maxyx()

    well_depth    := well_dimensions[0]
    well_width    := well_dimensions[1]
    vert_headroom := well_dimensions[2]

    well_left   := ( ( term_col / 2 ) - well_width )
    well_right  := ( well_left + ( well_width * 2 ) )
    well_bottom := ( vert_headroom + well_depth )

    // draw sides
    for row_height := vert_headroom ; row_height < well_bottom ; row_height ++ {
        stdscr.MovePrint( row_height , well_left  , " |" )
        stdscr.MovePrint( row_height , well_right , "| " )
    }

    for col_loc := well_left ; col_loc <= well_right ; col_loc ++ {
        stdscr.MovePrint( well_bottom , col_loc  , "=" )
    }

    stdscr.Refresh()

}

func new_block( well_dimensions , block_location []int ) {
    block_location[0] = well_dimensions[0]     // block_height
    block_location[1] = well_dimensions[1] / 2 // block_longitude
}

