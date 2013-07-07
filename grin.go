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
    well_dimensions[0] = 20  // well_depth
    well_dimensions[1] = 10  // well_width
    well_dimensions[2] = 5   // vert_headroom

    draw_border( stdscr , well_dimensions )

    debris_map := make( [][]int , well_dimensions[0] )
    for i := 0 ; i < well_dimensions[0] ; i++ {
        debris_row := make([]int, well_dimensions[1])
        debris_map[i] = debris_row
    }

    // starting block location
    block_location  := make( []int , 2 )
    new_block( stdscr , well_dimensions , block_location , debris_map )
    show_stats( stdscr , 1 , "block height  " , block_location[0] )
    // block_location[0] = well_dimensions[0] - 1 // block_height
    // block_location[1] = well_dimensions[1] / 2 // block_longitude

    for keep_going := true ; keep_going == true ; {

        show_stats( stdscr , 1 , "block height  " , block_location[0] )
        show_stats( stdscr , 2 , "block longtude" , block_location[1] )

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
        block_status := move_block( stdscr , well_dimensions , block_location , movement , debris_map )

        // new block?
        if block_status == 2 {
            debris_map[block_location[0]][block_location[1]] = 1
            nb_ret := new_block( stdscr , well_dimensions , block_location , debris_map )
            if nb_ret == 2 {
                keep_going = false
            }
        }

    }

	goncurses.End()

}


func show_stats( stdscr goncurses.Window , height int , show_text string , show_val int ) {

    bh_status := fmt.Sprintf( "%s : %02d     " , show_text , show_val )
    stdscr.MovePrint( height , 1  , bh_status )

}

func move_block( stdscr goncurses.Window , well_dimensions , block_location []int , operation string , debris_map [][]int) int {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    blocked := check_collisions( well_dimensions , block_location , debris_map , operation )

    if blocked == true {
        if operation == "dropone" {
            return 2
        } else {
            return 1
        }
    }

    draw_block( stdscr , well_dimensions , "erase" , block_location )

    retstat := 0
    switch {
        case operation == "left" :
            block_longitude--
        case operation == "right" :
            block_longitude++
        case operation == "dropone" :
            block_height--
        case operation == "drop" :
            block_height = sound_depth( block_location , debris_map )
            retstat = 2
    }

    block_location[0] = block_height
    block_location[1] = block_longitude

    draw_block( stdscr , well_dimensions , "draw" , block_location )

    return retstat

}

func check_collisions( well_dimensions , block_location []int , debris_map [][]int , operation string ) bool {

    block_height    := block_location[0]
    block_longitude := block_location[1]
    blocked := false
    switch {
        case operation == "left" :
            switch {
                case block_longitude == 0 :
                    blocked = true
                case debris_map[block_height][block_longitude - 1] == 1 :
                    blocked = true
            }
        case operation == "right" :
            switch {
                case block_location[1] == ( well_dimensions[1] - 1 ) :
                    blocked = true
                case debris_map[block_height][block_longitude + 1] == 1 :
                    blocked = true
            }
        case operation == "dropone" :
            switch {
                case block_location[0] == 0 :
                    blocked = true
                case debris_map[block_height - 1 ][block_longitude] == 1 :
                    blocked = true
            }
        case operation == "drop" :
            blocked = false // nothing to do here yet
    }

    return blocked
}

func sound_depth( block_location []int , debris_map [][]int ) int {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    for i := block_height ; i > 0  ; i-- {
        if debris_map[i][block_longitude] == 1 {
            return i + 1
        }
    }

    return 0
}

func draw_block( stdscr goncurses.Window , well_dimensions []int , operation string , block_location []int ) bool {

    // terminal size
    _, term_col := stdscr.Maxyx()

    well_left := ( ( term_col / 2 ) - well_dimensions[1] )

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

    well_left   := ( ( term_col / 2 ) - well_width ) - 2
    well_right  := well_left + ( well_width * 2 ) + 2
    well_bottom := vert_headroom + well_depth + 1

    // draw sides
    for row_height := vert_headroom ; row_height < well_bottom ; row_height ++ {
        stdscr.MovePrint( row_height , well_left  , " |" )
        stdscr.MovePrint( row_height , well_right , "| " )
    }

    for col_loc := ( well_left + 1 ) ; col_loc <= well_right ; col_loc ++ {
        stdscr.MovePrint( well_bottom , col_loc  , "=" )
    }

    stdscr.Refresh()

}

func new_block( stdscr goncurses.Window , well_dimensions , block_location []int , debris_map [][]int ) int {

    // show_stats( stdscr , 2 , "block ending loc" , block_location[1] )
    stdscr.GetChar()

    block_location[0] = well_dimensions[0] - 1 // block_height
    block_location[1] = well_dimensions[1] / 2 // block_longitude

    draw_debris( stdscr , well_dimensions , debris_map )

    return 0
}

func draw_debris( stdscr goncurses.Window , well_dimensions []int , debris_map [][]int ) {

    _, term_col := stdscr.Maxyx()
    vert_headroom := well_dimensions[2]

    // var well_width int
    for row := 0 ; row < len( debris_map ) ; row++ {
        for col := 0 ; col < len( debris_map[row] ) ; col++ {
            row_loc := vert_headroom + well_dimensions[0] - row
            col_loc := ( ( term_col / 2 ) - well_dimensions[1] ) + ( col * 2 )
            if debris_map[row][col] == 1 {
                stdscr.MovePrint( row_loc , col_loc  , "DD" )
            }
        }
    }

}
