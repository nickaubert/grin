package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "code.google.com/p/goncurses"
import "fmt"

func main() {

    // curses
	stdscr, _ := goncurses.Init()
	defer goncurses.End()

    // define well
    well_depth    := 20
    well_width    := 10
    vert_headroom := 5
    well_dimensions := make( []int , 3 )
    well_dimensions[0] = well_depth
    well_dimensions[1] = well_width
    well_dimensions[2] = vert_headroom

    draw_border( stdscr , well_dimensions )

    // tetromino
    t_size := 3
    tetronimo := make( [][]int , t_size )
    for i := 0 ; i < t_size ; i++ {
        tetro_row := make([]int, t_size)
        tetronimo[i] = tetro_row
    }

    // debris map
    debris_map := make( [][]int , well_depth + t_size )
    for i := 0 ; i < ( well_depth + t_size ) ; i++ {
        debris_row := make([]int, well_dimensions[1])
        debris_map[i] = debris_row
    }

    // starting block
    block_location  := make( []int , 2 )
    new_block( stdscr , well_dimensions , block_location , tetronimo , debris_map )
    show_stats( stdscr , 1 , "block height  " , block_location[0] )

    for keep_going := true ; keep_going == true ; {

        show_stats( stdscr , 1 , "block height  " , block_location[0] )
        show_stats( stdscr , 2 , "block longtude" , block_location[1] )
        show_stats( stdscr , 4 , "deb len       " , len( debris_map ) )


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
        block_status := move_block( stdscr , well_dimensions , block_location , movement , tetronimo , debris_map )

        // new block?
        if block_status == 2 {

            block_height    := block_location[0]
            block_longitude := block_location[1]
            for t_vert := range tetronimo {
                for t_horz := range tetronimo[t_vert] {
                    t_bit_vert := block_height    - t_vert
                    t_bit_horz := block_longitude + t_horz
                    if tetronimo[t_vert][t_horz] == 1 {
                        debris_map[t_bit_vert][t_bit_horz] = 1
                    }
                }
            }

            clear_debris( well_dimensions , debris_map , stdscr )
            nb_ret := new_block( stdscr , well_dimensions , block_location , tetronimo , debris_map )
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

func move_block( stdscr goncurses.Window , well_dimensions , block_location []int , operation string , tetronimo , debris_map [][]int) int {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    blocked := check_collisions( well_dimensions , block_location , tetronimo , debris_map , operation )

    if blocked == true {
        if operation == "dropone" {
            return 2
        } else {
            return 1
        }
    }

    draw_block( stdscr , well_dimensions , "erase" , block_location , tetronimo )

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

    draw_block( stdscr , well_dimensions , "draw" , block_location , tetronimo )

    return retstat

}

func check_collisions( well_dimensions , block_location []int , tetronimo , debris_map [][]int , operation string ) bool {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    for t_vert := range tetronimo {
        for t_horz := range tetronimo[t_vert] {

            t_bit_vert := block_height    - t_vert
            t_bit_horz := block_longitude + t_horz

            if tetronimo[t_vert][t_horz] == 1 {
                switch {
                    case operation == "left" :
                        switch {
                            case t_bit_horz == 0 :
                                return true
                            case debris_map[t_bit_vert][t_bit_horz - 1] == 1 :
                                return true
                        }
                    case operation == "right" :
                        switch {
                            case t_bit_horz == ( well_dimensions[1] - 1 ) :
                                return true
                            case debris_map[t_bit_vert][t_bit_horz + 1] == 1 :
                                return true
                        }
                    case operation == "dropone" :
                        switch {
                            case t_bit_vert == 0 :
                                return true
                            case debris_map[t_bit_vert - 1 ][t_bit_horz] == 1 :
                                return true
                        }
                    case operation == "drop" :
                        // nothing to do here yet
                }
            }
        }
    }

    return false
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

func draw_block( stdscr goncurses.Window , well_dimensions []int , operation string , block_location []int , tetronimo [][]int ) {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    block_paint := "XX"
    if operation == "erase" {
        block_paint = "  "
    }

    _, term_col := stdscr.Maxyx()
    well_bottom := well_dimensions[0] + well_dimensions[2]
    well_left := ( ( term_col / 2 ) - well_dimensions[1] )

    for t_vert := range tetronimo {
        for t_horz := range tetronimo[t_vert] {
            if tetronimo[t_vert][t_horz] == 1 {
                stdscr.MovePrint( ( well_bottom - block_height + t_vert ) , ( well_left + ( block_longitude * 2 ) + ( t_horz * 2 ) )  , block_paint )
            }
        }
    }

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

func new_block( stdscr goncurses.Window , well_dimensions , block_location []int , tetronimo , debris_map [][]int ) int {

    // show_stats( stdscr , 2 , "block ending loc" , block_location[1] )
    // stdscr.GetChar()

    block_location[0] = well_dimensions[0] - 1 // block_height
    block_location[1] = well_dimensions[1] / 2 // block_longitude

    /*
    // hardcode "+" block
    tetronimo[0][1] = 1
    tetronimo[1][0] = 1
    tetronimo[1][1] = 1
    tetronimo[1][2] = 1
    tetronimo[2][1] = 1
    */

    // hardcode "O" block
    tetronimo[0][0] = 1
    tetronimo[0][1] = 1
    tetronimo[1][0] = 1
    tetronimo[1][1] = 1

    block_height    := block_location[0]
    block_longitude := block_location[1]
    for t_vert := range tetronimo {
        for t_horz := range tetronimo[t_vert] {
            t_bit_vert := block_height    - t_vert
            t_bit_horz := block_longitude + t_horz
            if tetronimo[t_vert][t_horz] == 1 {
                /*
                debris_height := len( debris_map ) // testing
                show_stats( stdscr , 13 , "debris_height" , debris_height  )
                show_stats( stdscr , 14 , "debris_width" , len( debris_map[10])  )
                show_stats( stdscr , 14 , "debris_height" , len(debris_map)  )
                show_stats( stdscr , 15 , "debris_width_18" , len( debris_map[18])  )
                show_stats( stdscr , 16 , "debris_width_19" , len( debris_map[19])  )
                show_stats( stdscr , 17 , "debris_width_20" , len( debris_map[20])  )
                show_stats( stdscr , 18 , "t_bit_vert" , t_bit_vert  )
                show_stats( stdscr , 19 , "t_bit_horz" , t_bit_horz  )
                stdscr.GetChar()
                */
                if debris_map[t_bit_vert][t_bit_horz] == 1 {
                    return 2 // game over!
                }
            }
        }
    }

    draw_debris( stdscr , well_dimensions , debris_map )

    return 0
}

func draw_debris( stdscr goncurses.Window , well_dimensions []int , debris_map [][]int ) {

    _, term_col := stdscr.Maxyx()
    vert_headroom := well_dimensions[2]

    // var well_width int
    for row := range debris_map {
        for col := range debris_map[row] {
            row_loc := vert_headroom + well_dimensions[0] - row
            col_loc := ( ( term_col / 2 ) - well_dimensions[1] ) + ( col * 2 )
            if debris_map[row][col] == 1 {
                stdscr.MovePrint( row_loc , col_loc  , "DD" )
            } else {
                stdscr.MovePrint( row_loc , col_loc  , "  " )
            }
        }
    }

}

// stdscr for debugging
func clear_debris ( well_dimensions []int , debris_map [][]int , stdscr goncurses.Window ) {

    deb_height := len( debris_map )
    well_width := well_dimensions[1]

    clear_rows := make( []int , deb_height )

    // do_refresh := false
    delete_rows := 0
    for d_vert := range debris_map {
        d_count := 0
        for d_horz := range debris_map[d_vert] {
            if debris_map[d_vert][d_horz] == 1 {
                d_count++
            }
        }
        if d_count == well_width {
            delete_rows++
            clear_rows[d_vert] = 1
        }
    }

    // return here if no clear rows
    if delete_rows == 0 {
        return
    }

    new_debris := make( [][]int , len(debris_map) )
    new_rows := 0
    for d_vert := range debris_map {
        if clear_rows[d_vert] == 1 {
            // do nothing
        } else {
            new_debris[new_rows] = debris_map[d_vert]
            new_rows++
        }
    }

    show_stats( stdscr , 7 , "ndeb len " , len(new_debris) )
    for i := ( len(debris_map) - delete_rows ) ; i < len(debris_map) ; i++ {
        // TODO: make sure array has length!
        fresh_row := make( []int , well_width )
        new_debris[i] = fresh_row
        show_stats( stdscr , 7 , "adding to " , i )
    }

    for this_row := range new_debris {
        debris_map[this_row] = new_debris[this_row]
    }

    for this_row := range debris_map {
        offset_show  := 8
        show_stats( stdscr , offset_show + this_row , "array len " , len( debris_map[this_row] )  )
    }
    stdscr.GetChar()

    /*
    for i := 0 ; i < delete_rows ; i++ {
        for d_vert := range debris_map {
            if clear_rows[d_vert] == 1 {
                show_stats( stdscr , 5 , "old len " , len( debris_map)  )
                debris_map = append(debris_map[:d_vert], debris_map[d_vert+1:]...)
                // deleted_rows++
                show_stats( stdscr , 6 , "deleting row" , d_vert  )
                show_stats( stdscr , 7 , "new len " , len( debris_map)  )
                stdscr.GetChar()
            }
        }
        fresh_row := make( []int , well_width )
        debris_map = append( debris_map , fresh_row )
    }
    */

    /*
    for i := 0 ; i < deleted_rows ; i++ {
        fresh_row := make( []int , well_width )
        debris_map = append( debris_map , fresh_row )
    }
    */

    /* Manual Method
    down_rows := 0
    for d_vert := range debris_map {

        if clear_rows[d_vert] == 1 {
            down_rows++
            show_stats( stdscr , 20 , "clearing row" , d_vert  )
            stdscr.GetChar()
        }

        if ( d_vert + down_rows ) <= ( len( debris_map ) - 1 ) {
          debris_map[d_vert] = debris_map[d_vert+down_rows]
        } else {
            fresh_row := make( []int , well_width )
            debris_map[d_vert] = fresh_row
        }

    }
    */

}

