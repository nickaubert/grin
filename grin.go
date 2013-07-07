package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import gc "code.google.com/p/goncurses"
import "fmt"
import "math/rand"
import "time"
import "runtime"

func main() {

    // curses
	stdscr, _ := gc.Init()
	defer gc.End()

    // curses colors
    gc.StartColor()
    gc.InitPair(1, gc.C_BLACK, gc.C_BLUE)
    gc.InitPair(2, gc.C_BLACK, gc.C_YELLOW)
    gc.InitPair(3, gc.C_BLACK, gc.C_MAGENTA)
    gc.InitPair(4, gc.C_BLACK, gc.C_WHITE)
    gc.InitPair(5, gc.C_BLACK, gc.C_GREEN)
    gc.InitPair(6, gc.C_BLACK, gc.C_CYAN)
    gc.InitPair(7, gc.C_BLACK, gc.C_RED)
    gc.InitPair(8, gc.C_BLACK, gc.C_BLACK)

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
    t_size := 4
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
    block_id := new_block( stdscr , well_dimensions , block_location , tetronimo , debris_map , t_size )

    // keyboard channel
    ck := make(chan int )

    // timer channel
    ct := make(chan int )

    go keys_in( stdscr , ck )
    go t_timer ( ct )

    for keep_going := true ; keep_going == true ; {

        var somechar int
        hithere := 0
        select {
        case somechar = <-ck:
            go keys_in( stdscr , ck )
        case somechar = <-ct:
            go t_timer ( ct )
            hithere = 1
        }
        show_stats( stdscr , 2 , "hithere    " , hithere )

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
            case somechar == 107 : // k
                movement = "rotate"
            case somechar == 32 :  // [space]
                movement = "drop"
            case somechar == 200 :  // TESTING
                show_stats( stdscr , 9 , "drop    " , somechar )
        }

        if keep_going == false {
            break
        }

        // move block 
        block_status := move_block( stdscr , well_dimensions , block_location , movement , block_id , tetronimo , debris_map )

        // new block?
        if block_status == 2 {

            block_height    := block_location[0]
            block_longitude := block_location[1]
            for t_vert := range tetronimo {
                for t_horz := range tetronimo[t_vert] {
                    t_bit_vert := block_height    - t_vert
                    t_bit_horz := block_longitude + t_horz
                    if tetronimo[t_vert][t_horz] == 1 {
                        debris_map[t_bit_vert][t_bit_horz] = block_id + 1
                    }
                }
            }

            clear_debris( well_dimensions , debris_map , stdscr )
            // nb_ret := new_block( stdscr , well_dimensions , block_location , tetronimo , debris_map , t_size )
            block_id = new_block( stdscr , well_dimensions , block_location , tetronimo , debris_map , t_size )
            if block_id == 8 {
                keep_going = false
            }
        }
        show_stats( stdscr , 1 , "goroutines    " , runtime.NumGoroutine() )
        stdscr.Refresh()


    }

	gc.End()

}


func show_stats( stdscr gc.Window , height int , show_text string , show_val int ) {

    bh_status := fmt.Sprintf( "%s : %02d     " , show_text , show_val )
    stdscr.MovePrint( height , 1  , bh_status )

}

func move_block( stdscr gc.Window , well_dimensions , block_location []int , operation string , block_id int , tetronimo , debris_map [][]int) int {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    blocked := check_collisions( well_dimensions , block_location , tetronimo , debris_map , operation , stdscr )

    if blocked == true {
        if operation == "dropone" {
            return 2
        } else {
            return 1
        }
    }

    draw_block( stdscr , well_dimensions , "erase" , block_location , tetronimo , block_id )

    retstat := 0
    switch {
        case operation == "left" :
            block_longitude--
        case operation == "right" :
            block_longitude++
        case operation == "rotate" :
            rotate_tetronimo( tetronimo )
        case operation == "dropone" :
            block_height--
        case operation == "drop" :
            block_height = sound_depth( block_location , debris_map )
            retstat = 2
    }

    block_location[0] = block_height
    block_location[1] = block_longitude

    draw_block( stdscr , well_dimensions , "draw" , block_location , tetronimo , block_id )

    return retstat

}

func check_collisions( well_dimensions , block_location []int , tetronimo , debris_map [][]int , operation string , stdscr gc.Window ) bool {

    ghost_height    := block_location[0]
    ghost_longitude := block_location[1]

    ghost_tetro := make( [][]int , len(tetronimo) )
    for row := 0 ; row < len(tetronimo) ; row++ {
        tetro_row := make([]int, len(tetronimo[0]))
        ghost_tetro[row] = tetro_row
        for col := 0 ; col < len(tetronimo[0]) ; col++ {
            ghost_tetro[row][col] = tetronimo[row][col]
        }
    }

    switch {
        case operation == "left" :
            ghost_longitude--
        case operation == "right" :
            ghost_longitude++
        case operation == "rotate" :
            rotate_tetronimo( ghost_tetro )
        case operation == "dropone" :
            ghost_height--
        case operation == "drop" :
            ghost_height = sound_depth( block_location , debris_map )
            // retstat = 2
    }

    for t_vert := range ghost_tetro {
        for t_horz := range ghost_tetro[t_vert] {

            t_bit_vert := ghost_height    - t_vert
            t_bit_horz := ghost_longitude + t_horz

            if ghost_tetro[t_vert][t_horz] == 1 {
                if t_bit_horz < 0 {
                    return true
                }
                if t_bit_horz >= well_dimensions[1] {
                    return true
                }
                if t_bit_vert < 0 {
                    return true
                }
                if debris_map[t_bit_vert][t_bit_horz] > 0 {
                    return true
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
        if debris_map[i][block_longitude] > 0 {
            return i + 1
        }
    }

    return 0
}

func draw_block( stdscr gc.Window , well_dimensions []int , operation string , block_location []int , tetronimo [][]int , block_id int ) {

    block_height    := block_location[0]
    block_longitude := block_location[1]

    block_paint := "  "
    // color := 2
    color := block_id + 1

    show_stats( stdscr , 6 , "block_id    " , block_id )

    if operation == "erase" {
        color = 8
    }

    _, term_col := stdscr.Maxyx()
    well_bottom := well_dimensions[0] + well_dimensions[2]
    well_left := ( ( term_col / 2 ) - well_dimensions[1] )

    for t_vert := range tetronimo {
        for t_horz := range tetronimo[t_vert] {
            if tetronimo[t_vert][t_horz] == 1 {
                stdscr.ColorOn(byte(color))
                stdscr.MovePrint( ( well_bottom - block_height + t_vert ) , ( well_left + ( block_longitude * 2 ) + ( t_horz * 2 ) )  , block_paint )
                stdscr.ColorOff(byte(color))
            }
        }
    }

}

func draw_border( stdscr gc.Window , well_dimensions []int ) {

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

        stdscr.MovePrint( row_height , well_right , "| " )

        stdscr.MovePrint( row_height , well_left  , " |" )

    }

    for col_loc := ( well_left + 1 ) ; col_loc <= well_right ; col_loc ++ {
        stdscr.MovePrint( well_bottom , col_loc  , "=" )
    }

    stdscr.Refresh()

}

func new_block( stdscr gc.Window , well_dimensions , block_location []int , tetronimo , debris_map [][]int , t_size int ) int {

    block_location[0] = well_dimensions[0] - 1 // block_height
    block_location[1] = well_dimensions[1] / 2 // block_longitude

    // show_stats( stdscr , 8 , "random" , rand_tetro )
    block_id := rand_block( tetronimo , t_size )

    block_height    := block_location[0]
    block_longitude := block_location[1]
    for t_vert := range tetronimo {
        for t_horz := range tetronimo[t_vert] {
            t_bit_vert := block_height    - t_vert
            t_bit_horz := block_longitude + t_horz
            if tetronimo[t_vert][t_horz] == 1 {
                if debris_map[t_bit_vert][t_bit_horz] > 0 {
                    return 8 // game over!
                }
            }
        }
    }

    draw_debris( stdscr , well_dimensions , debris_map )

    return block_id
}

func draw_debris( stdscr gc.Window , well_dimensions []int , debris_map [][]int ) {

    _, term_col := stdscr.Maxyx()
    vert_headroom := well_dimensions[2]

    // var well_width int
    for row := range debris_map {
        for col := range debris_map[row] {
            row_loc := vert_headroom + well_dimensions[0] - row
            col_loc := ( ( term_col / 2 ) - well_dimensions[1] ) + ( col * 2 )
            if debris_map[row][col] > 0 {
                color := debris_map[row][col]
                // stdscr.MovePrint( row_loc , col_loc  , "DD" )
                stdscr.ColorOn(byte(color))
                stdscr.MovePrint( row_loc , col_loc  , "  " )
                stdscr.ColorOff(byte(color))
            } else {
                stdscr.MovePrint( row_loc , col_loc  , "  " )
            }
        }
    }

}

// stdscr for debugging
func clear_debris ( well_dimensions []int , debris_map [][]int , stdscr gc.Window ) {

    deb_height := len( debris_map )
    well_width := well_dimensions[1]

    clear_rows := make( []int , deb_height )

    // do_refresh := false
    delete_rows := 0
    for d_vert := range debris_map {
        d_count := 0
        for d_horz := range debris_map[d_vert] {
            if debris_map[d_vert][d_horz] > 0 {
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

    for i := ( len(debris_map) - delete_rows ) ; i < len(debris_map) ; i++ {
        fresh_row := make( []int , well_width )
        new_debris[i] = fresh_row
    }

    for this_row := range new_debris {
        debris_map[this_row] = new_debris[this_row]
    }

}

func rand_block( tetronimo [][]int , t_size int ) int {

    set_count := 7

    tetronimo_set := make( [][][]int , set_count )
    for set_num := 0 ; set_num < set_count ; set_num++ {
        tetro_row := make([][]int, t_size)
        for i := 0 ; i < t_size ; i++ {
            tetro_col := make([]int, t_size)
            tetro_row[i] = tetro_col
        }
        tetronimo_set[set_num] = tetro_row
    }

    // define "O" block
    tetronimo_set[0][1][1] = 1
    tetronimo_set[0][1][2] = 1
    tetronimo_set[0][2][1] = 1
    tetronimo_set[0][2][2] = 1

    // define "T" block
    tetronimo_set[1][0][0] = 1
    tetronimo_set[1][0][1] = 1
    tetronimo_set[1][0][2] = 1
    tetronimo_set[1][1][1] = 1

    // define "L" block
    tetronimo_set[2][0][0] = 1
    tetronimo_set[2][1][0] = 1
    tetronimo_set[2][2][0] = 1
    tetronimo_set[2][2][1] = 1

    // define "J" block
    tetronimo_set[3][0][1] = 1
    tetronimo_set[3][1][1] = 1
    tetronimo_set[3][2][0] = 1
    tetronimo_set[3][2][1] = 1

    // define "S" block
    tetronimo_set[4][0][0] = 1
    tetronimo_set[4][1][0] = 1
    tetronimo_set[4][1][1] = 1
    tetronimo_set[4][2][1] = 1

    // define "Z" block
    tetronimo_set[5][0][1] = 1
    tetronimo_set[5][1][0] = 1
    tetronimo_set[5][1][1] = 1
    tetronimo_set[5][2][0] = 1

    // define "I" block
    tetronimo_set[6][0][1] = 1
    tetronimo_set[6][1][1] = 1
    tetronimo_set[6][2][1] = 1
    tetronimo_set[6][3][1] = 1

    rand.Seed(time.Now().Unix())
    rand_tetro := rand.Intn(set_count)

    for row := range tetronimo {
        for col := range tetronimo[row] {
            tetronimo[row][col] = tetronimo_set[rand_tetro][row][col]
        }
    }

    return rand_tetro

}

func rotate_tetronimo ( tetronimo [][]int ) {

    hold_tetro := make( [][]int , len(tetronimo) )
    for row := 0 ; row < len(tetronimo) ; row++ {
        tetro_row := make([]int, len(tetronimo[0]))
        hold_tetro[row] = tetro_row
        for col := 0 ; col < len(tetronimo[0]) ; col++ {
            hold_tetro[row][col] = tetronimo[row][col]
        }
    }

    // stupid hardcode rotate
    tetronimo[0][0] = hold_tetro[0][3]
    tetronimo[0][1] = hold_tetro[1][3]
    tetronimo[0][2] = hold_tetro[2][3]
    tetronimo[0][3] = hold_tetro[3][3]
    tetronimo[1][0] = hold_tetro[0][2]
    tetronimo[1][1] = hold_tetro[1][2]
    tetronimo[1][2] = hold_tetro[2][2]
    tetronimo[1][3] = hold_tetro[3][2]
    tetronimo[2][0] = hold_tetro[0][1]
    tetronimo[2][1] = hold_tetro[1][1]
    tetronimo[2][2] = hold_tetro[2][1]
    tetronimo[2][3] = hold_tetro[3][1]
    tetronimo[3][0] = hold_tetro[0][0]
    tetronimo[3][1] = hold_tetro[1][0]
    tetronimo[3][2] = hold_tetro[2][0]
    tetronimo[3][3] = hold_tetro[3][0]

}

func keys_in ( stdscr gc.Window , ck chan int ) {
    somechar := int( stdscr.GetChar() )
    ck <- somechar
}

func t_timer ( ct chan int ) {
    time.Sleep(1000 * time.Millisecond)
    ct <- 110
}
