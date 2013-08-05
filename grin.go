package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "fmt"
import "math/rand"
import "time"
import "os"

import tb "github.com/nsf/termbox-go"

/*
	TODO:
		Cleanup functions, data objects?
		Print score, stats on exit
		Change tetronimo shape from grid to vector?
	Improvements:
		Adjustible well size and tetronimo set
		Random debris map at start
		Two players?
	Done!
		Fix rotate: when rotate, move top left of grid
		Hard drop
		Draw well
		Softcode well size
		Speedup
		Screen cleanup
		Keep score, stats
			Tint:
				1 point per level for each tetronimo
				1 point per level for each distance dropped
				0 points for clearing rows
		*** CONVERT FROM goncurses to termbox
*/

type Tetronimo struct {
	shape [][]tb.Attribute
	height int
	longitude int
}

type Well struct {
	debris_map [][]tb.Attribute
}

func main() {

	// init termbox
	err := tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	// tetromino
	t_size := 4
	tetronimo := new( Tetronimo )
	set_tetronimo( tetronimo , t_size )

	// define well
	well_depth := 20
	well_width := 10
	well := new( Well )
	set_well( well , well_depth , well_width )
	draw_border(well)

	// score
	//  0 tetronimo count
	//  1 block count (ie. 4 per tetronimo)
	//  2 score count
	//  3 rows deleted
	score := make([][]int, 2)
	score_thing := make([]int, 4)
	score_tetros := make([]int, 7) // 7 different tetronimos
	score[0] = score_thing
	score[1] = score_tetros

	// starting block
	block_id := new_block( well , tetronimo , score )

	// keyboard channel
	ck := make(chan rune)

	// timer channel
	ct := make(chan rune)

	speed := 1
	go keys_in(ck)
	go t_timer(ct, speed)

	for keep_going := true; keep_going == true; {

		var action string
		var somechar rune
		select {
		case somechar = <-ct:
			action = "timeoff"
		case somechar = <-ck:
			action = "keyboard"
		}

		operation := "hold"
		pause := false
		switch {
		case somechar == 113: // q
			keep_going = false
		case somechar == 106: // j
			operation = "left"
		case somechar == 108: // l
			operation = "right"
		case somechar == 110: // n
			operation = "dropone"
		case somechar == 112: // p
			pause = true
		case somechar == 107: // k
			operation = "rotate"
		case somechar == 0: // 32: // [space]
			operation = "harddrop"
		case somechar == 200: // TESTING
		}

		if keep_going == false {
			break
		}

		// pause
		if pause == true {
			_ = tb.PollEvent().Ch
		}

		// move block
		block_status := 0

		blocked := check_collisions( well,  tetronimo , operation)

		if blocked == true {
			if operation == "dropone" {
				block_status = 2
			} else {
				block_status = 1
			}
		} else {

			draw_block(well, "erase", tetronimo)

			block_status = 0
			switch {
			case operation == "left":
				tetronimo.longitude--
			case operation == "right":
				tetronimo.longitude++
			case operation == "rotate":
				rotate_tetronimo(tetronimo)
			case operation == "dropone":
				tetronimo.height--
			case operation == "harddrop":
				sound_depth( tetronimo, well )
				block_status = 2
			}

			draw_block(well, "draw", tetronimo)

		}

		// new block?
		if block_status == 2 {

			// new struct method
			for t_vert := range tetronimo.shape {
				for t_horz := range tetronimo.shape[t_vert] {
					t_bit_vert := tetronimo.height - t_vert
					t_bit_horz := tetronimo.longitude + t_horz
					if tetronimo.shape[t_vert][t_horz] > 0 {
						well.debris_map[t_bit_vert][t_bit_horz] = tetronimo.shape[t_vert][t_horz]
					}
				}
			}

			clear_debris( well, score)
			block_id = new_block(well, tetronimo , score )
			draw_debris(well)
			if block_id == 8 {
				keep_going = false
			}

		}

		// speedup
		if score[0][2] > 0 {
			speed = int(score[0][2] / 10)
			if speed == 0 {
				speed = 1
			}
		}

		show_stats(4, "tetronimos: ", score[0][0])
		show_stats(5, "blocks    : ", score[0][1])
		show_stats(6, "rows      : ", score[0][2])

		show_stats(8, "tet O     : ", score[1][0])
		show_stats(9, "tet T     : ", score[1][1])
		show_stats(10, "tet L     : ", score[1][2])
		show_stats(11, "tet J     : ", score[1][3])
		show_stats(12, "tet S     : ", score[1][4])
		show_stats(13, "tet Z     : ", score[1][5])
		show_stats(14, "tet I     : ", score[1][6])

		show_stats(16, "speed     : ", speed)

		// someint := int(somechar)
		// show_stats(18, "keypress  : ", someint)

		tb.Flush()

		switch {
		case action == "timeoff":
			go t_timer(ct, speed)
		case action == "keyboard":
			go keys_in(ck)
		}

	}

	tb.Close()
	fmt.Print("tetronimos: ", score[0][0], "\n")
	fmt.Print("blocks    : ", score[0][1], "\n")
	fmt.Print("rows      : ", score[0][2], "\n")
	fmt.Print("speed     : ", speed, "\n")
	fmt.Print("tet O     : ", score[1][0], "\n")
	fmt.Print("tet T     : ", score[1][1], "\n")
	fmt.Print("tet L     : ", score[1][2], "\n")
	fmt.Print("tet J     : ", score[1][3], "\n")
	fmt.Print("tet S     : ", score[1][4], "\n")
	fmt.Print("tet Z     : ", score[1][5], "\n")
	fmt.Print("tet I     : ", score[1][6], "\n")
	fmt.Print("Game over\n")
	os.Exit(0)

}

func debug_stats(height int, show_text string, show_val int) {

	bh_status := fmt.Sprintf("%s : %d     ", show_text, show_val)
	print_tb(0, height, 0, 0, bh_status)
	_ = tb.PollEvent().Ch

}

func show_stats(height int, show_text string, show_val int) {

	bh_status := fmt.Sprintf("%s : %d     ", show_text, show_val)
	print_tb(0, height, 0, 0, bh_status)

}

func check_collisions( well *Well , this_tetronimo *Tetronimo , operation string) bool {

	ghost_tetronimo := new( Tetronimo )
	set_tetronimo( ghost_tetronimo , len(this_tetronimo.shape) )
	clone_tetronimo( this_tetronimo , ghost_tetronimo )

	switch {
	case operation == "left":
		ghost_tetronimo.longitude--
	case operation == "right":
		ghost_tetronimo.longitude++
	case operation == "rotate":
		rotate_tetronimo(ghost_tetronimo)
	case operation == "dropone":
		ghost_tetronimo.height--
	case operation == "harddrop":
		return false
	}

	for t_vert := range ghost_tetronimo.shape {
		for t_horz := range ghost_tetronimo.shape[t_vert] {

			t_bit_vert := ghost_tetronimo.height - t_vert
			t_bit_horz := ghost_tetronimo.longitude + t_horz

			if ghost_tetronimo.shape[t_vert][t_horz] > 0 {
				if t_bit_horz < 0 {
					return true
				}
				if t_bit_horz >= len(well.debris_map[0]) {
					return true
				}
				if t_bit_vert < 0 {
					return true
				}
				if well.debris_map[t_bit_vert][t_bit_horz] > 0 {
					return true
				}
			}
		}
	}

	return false
}

func sound_depth( this_tetronimo *Tetronimo, well *Well) {

	ghost_tetronimo := new( Tetronimo )
	set_tetronimo( ghost_tetronimo , len(this_tetronimo.shape) )
	clone_tetronimo( this_tetronimo , ghost_tetronimo )

	for ghost_height := ghost_tetronimo.height; ghost_height >= 0; ghost_height-- {
		ghost_tetronimo.height = ghost_height
		blocked := check_collisions(well , ghost_tetronimo, "dropone")
		if blocked == true {
			this_tetronimo.height = ghost_height
			return
		}
	}

}

func draw_block( well *Well, operation string, this_tetronimo *Tetronimo ) {

	well_depth := len(well.debris_map)

	term_col, term_row := tb.Size()
	vert_headroom := int((term_row-well_depth)/2) - 1

	well_bottom := len(well.debris_map) + vert_headroom
	well_left := ((term_col / 2) - len(well.debris_map[0]))

	for t_vert := range this_tetronimo.shape {
		for t_horz := range this_tetronimo.shape[t_vert] {
			if this_tetronimo.shape[t_vert][t_horz] > 0 {
				color := tb.ColorDefault
				if operation == "draw" {
					color = this_tetronimo.shape[t_vert][t_horz]
				}
				height := (well_bottom - this_tetronimo.height + t_vert)
				longitude := (well_left + (this_tetronimo.longitude * 2) + (t_horz * 2))
				tb.SetCell(longitude, height, 0, 0, color)
				tb.SetCell(longitude+1, height, 0, 0, color)
			}
		}
	}

}

func draw_border( well *Well ) {

	// terminal size
	term_col, term_row := tb.Size()

	well_depth := len(well.debris_map)
	well_width := len(well.debris_map[0])

	if well_depth+1 >= term_row {
		error_out("too short!\n")
	}

	vert_headroom := int((term_row-well_depth)/2) - 1

	well_left := ((term_col / 2) - well_width) - 2
	well_right := well_left + (well_width * 2) + 2
	well_bottom := vert_headroom + well_depth + 1

	vline := rune(0x2502)
	bleft := rune(0x2514)
	bright := rune(0x2518)
	hline := rune(0x2500)

	// draw well sides
	for row_height := vert_headroom; row_height < well_bottom; row_height++ {
		tb.SetCell(well_right, row_height, vline, 0, 0)
		tb.SetCell(well_left+1, row_height, vline, 0, 0)
	}

	// draw well bottom
	for col_loc := (well_left + 2); col_loc < well_right; col_loc++ {
		tb.SetCell(col_loc, well_bottom, hline, 0, 0)
	}

	// draw well corners
	tb.SetCell(well_left+1, well_bottom, bleft, 0, 0)
	tb.SetCell(well_right, well_bottom, bright, 0, 0)

	tb.Flush()

}

func new_block( well *Well, tetronimo *Tetronimo , score [][]int ) int {

	tetronimo.height    = len(well.debris_map) - 1
	tetronimo.longitude = len(well.debris_map[0]) / 2

	rand_block( tetronimo , score )

	for t_vert := range tetronimo.shape {
		for t_horz := range tetronimo.shape[t_vert] {
			t_bit_vert := tetronimo.height - t_vert
			t_bit_horz := tetronimo.longitude + t_horz
			if tetronimo.shape[t_vert][t_horz] > 0 {
				if well.debris_map[t_bit_vert][t_bit_horz] > 0 {
					return 8 // game over!
				}
			}
		}
	}

	return 0

}

func draw_debris(well *Well) {

	term_col, term_row := tb.Size()
	well_depth := len(well.debris_map)
	vert_headroom := int((term_row-well_depth)/2) - 1

	for row := range well.debris_map {
		for col := range well.debris_map[row] {
			row_loc := vert_headroom + len(well.debris_map) - row
			col_loc := ((term_col / 2) - len(well.debris_map[0])) + (col * 2)
			if well.debris_map[row][col] > 0 {
				color := well.debris_map[row][col]
				tb.SetCell(col_loc, row_loc, 0, 0, color)
				tb.SetCell(col_loc+1, row_loc, 0, 0, color)
			} else {
				tb.SetCell(col_loc, row_loc, 0, 0, tb.ColorDefault)
				tb.SetCell(col_loc+1, row_loc, 0, 0, tb.ColorDefault)
			}
		}
	}

}
func clear_debris( well *Well , score [][]int) {

	deb_height := len(well.debris_map)
	well_width := len(well.debris_map[0])

	clear_rows := make([]int, deb_height)

	delete_rows := 0
	for d_vert := range well.debris_map {
		d_count := 0
		for d_horz := range well.debris_map[d_vert] {
			if well.debris_map[d_vert][d_horz] > 0 {
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

	score[0][2] += delete_rows

	new_debris := make([][]tb.Attribute, len(well.debris_map))
	new_rows := 0
	for d_vert := range well.debris_map {
		if clear_rows[d_vert] == 1 {
			// do nothing
		} else {
			new_debris[new_rows] = well.debris_map[d_vert]
			new_rows++
		}
	}

	for i := (len(well.debris_map) - delete_rows); i < len(well.debris_map); i++ {
		fresh_row := make([]tb.Attribute, well_width)
		new_debris[i] = fresh_row
	}

	for this_row := range new_debris {
		well.debris_map[this_row] = new_debris[this_row]
	}

}

func rand_block( this_tetronimo *Tetronimo, score [][]int ) {

	set_count := 7

	tetronimo_set := make([][][]tb.Attribute, set_count+1)
	for set_num := 0; set_num <= set_count; set_num++ {

		tetro_def := make([][]tb.Attribute, 2)
		tetronimo_set[set_num] = tetro_def

		tetro_row := make([][]tb.Attribute, len(this_tetronimo.shape) )
		for i := 0; i < len(this_tetronimo.shape) ; i++ {
			tetro_col := make([]tb.Attribute, len(this_tetronimo.shape) )
			tetro_row[i] = tetro_col
		}
		tetronimo_set[set_num] = tetro_row
	}

	// there is no tetronimo_set[0]

	// define "O" block
	tetronimo_set[1][0][0] = tb.ColorBlue
	tetronimo_set[1][0][1] = tb.ColorBlue
	tetronimo_set[1][1][0] = tb.ColorBlue
	tetronimo_set[1][1][1] = tb.ColorBlue

	// define "T" block
	tetronimo_set[2][0][0] = tb.ColorYellow
	tetronimo_set[2][0][1] = tb.ColorYellow
	tetronimo_set[2][0][2] = tb.ColorYellow
	tetronimo_set[2][1][1] = tb.ColorYellow

	// define "L" block
	tetronimo_set[3][0][0] = tb.ColorMagenta
	tetronimo_set[3][1][0] = tb.ColorMagenta
	tetronimo_set[3][2][0] = tb.ColorMagenta
	tetronimo_set[3][2][1] = tb.ColorMagenta

	// define "J" block
	tetronimo_set[4][0][1] = tb.ColorWhite
	tetronimo_set[4][1][1] = tb.ColorWhite
	tetronimo_set[4][2][0] = tb.ColorWhite
	tetronimo_set[4][2][1] = tb.ColorWhite

	// define "S" block
	tetronimo_set[5][0][0] = tb.ColorGreen
	tetronimo_set[5][1][0] = tb.ColorGreen
	tetronimo_set[5][1][1] = tb.ColorGreen
	tetronimo_set[5][2][1] = tb.ColorGreen

	// define "Z" block
	tetronimo_set[6][0][1] = tb.ColorCyan
	tetronimo_set[6][1][0] = tb.ColorCyan
	tetronimo_set[6][1][1] = tb.ColorCyan
	tetronimo_set[6][2][0] = tb.ColorCyan

	// define "I" block
	tetronimo_set[7][0][0] = tb.ColorRed
	tetronimo_set[7][1][0] = tb.ColorRed
	tetronimo_set[7][2][0] = tb.ColorRed
	tetronimo_set[7][3][0] = tb.ColorRed

	rand.Seed(time.Now().Unix())
	rand_tetro := rand.Intn(set_count)
	// rand_tetro := 6  // TESTING i block
	// rand_tetro := 0  // TESTING o block

	b_count := 0
	for row := range this_tetronimo.shape {
		for col := range this_tetronimo.shape[row] {
			this_block := tetronimo_set[rand_tetro+1][row][col]
			if this_block > 0 {
				b_count++
			}
			this_tetronimo.shape[row][col] = this_block
		}
	}

	score[0][0] += 1
	score[0][1] += b_count
	score[1][rand_tetro] += 1

}

func rotate_tetronimo( this_tetronimo *Tetronimo ) {

	hold_tetronimo := new( Tetronimo )
	set_tetronimo( hold_tetronimo , len(this_tetronimo.shape) )
	clone_tetronimo( this_tetronimo , hold_tetronimo )

	// rotate
	tl_tetronimo := new( Tetronimo )
	set_tetronimo( tl_tetronimo , len(this_tetronimo.shape) )
	for row := range hold_tetronimo.shape {
		rotated_col := (len(hold_tetronimo.shape) - 1) - row
		for col := range hold_tetronimo.shape[row] {
			rotated_row := col
			tl_tetronimo.shape[row][col] = hold_tetronimo.shape[rotated_row][rotated_col]
		}
	}

	// push toward top left corner of grid
	row_top := 0
	col_left := 0
	row_offset := 0
	col_offset := 0
	for row := range tl_tetronimo.shape {
		for _, col_val := range tl_tetronimo.shape[row] {
			row_top += int(col_val)
		}
		if row_top == 0 {
			row_offset += 1
		}
		col_left += int(tl_tetronimo.shape[row][0])
	}

	if col_left == 0 {
		col_offset = 1
	}
	/*
		if top_row == 0 {
			row_offset = 1
		}
	*/
	if row_offset > 2 {
		row_offset = 2
	}

	for row := range tl_tetronimo.shape {
		this_row := row - row_offset
		if this_row < 0 {
			for col := range tl_tetronimo.shape[row] {
				this_tetronimo.shape[len(tl_tetronimo.shape)+this_row][col] = 0
			}
		} else {
			for col := range tl_tetronimo.shape[this_row] {
				this_col := col - col_offset
				if this_col < 0 {
					this_tetronimo.shape[this_row][len(tl_tetronimo.shape[row])-col_offset] = 0
				} else {
					this_tetronimo.shape[this_row][this_col] = tl_tetronimo.shape[row][col]
				}
			}
		}
	}

}

func keys_in(ck chan rune) {
	char := tb.PollEvent().Ch
	ck <- char
}

func t_timer(ct chan rune, speed int) {
	mseconds := time.Duration(1000 / speed)
	time.Sleep(mseconds * time.Millisecond)
	ct <- rune(110)
}

func error_out(message string) {

	fmt.Print(message)
	os.Exit(1)

}

func print_tb(x, y int, fg, bg tb.Attribute, msg string) {
	for _, c := range msg {
		tb.SetCell(x, y, c, fg, bg)
		x++
	}
	tb.Flush()
}

func clone_tetronimo( orig_tetronimo , new_tetronimo *Tetronimo ) {

	new_tetronimo.height = orig_tetronimo.height
	new_tetronimo.longitude = orig_tetronimo.longitude

	for row := 0; row < len(orig_tetronimo.shape); row++ {
		for col := 0; col < len(orig_tetronimo.shape[0]); col++ {
			new_tetronimo.shape[row][col] = orig_tetronimo.shape[row][col]
		}
	}

}

func set_tetronimo( tetronimo *Tetronimo , t_size int ) {

	this_tetronimo := make([][]tb.Attribute, t_size)
	for i := 0; i < t_size; i++ {
		tetro_row := make([]tb.Attribute, t_size)
		this_tetronimo[i] = tetro_row
	}

	tetronimo.height = 3
	tetronimo.longitude = 4
	tetronimo.shape = this_tetronimo

}

func set_well( well *Well , well_depth , well_width int ) {

	this_well := make([][]tb.Attribute, well_depth )
	for i := 0; i < well_depth; i++ {
		this_row := make([]tb.Attribute, well_width )
		this_well[i] = this_row
	}

	well.debris_map = this_well

}
