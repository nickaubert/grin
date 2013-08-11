package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "fmt"
import "math/rand"
import "time"
import "os"

import tb "github.com/nsf/termbox-go"

/*
	Improvements:
		High scores in sqlite?
		arrow keys
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
		Print score, stats on exit
		*** CONVERT FROM goncurses to termbox
		Cleanup functions, data objects (can always refactor)
		Phantom blocks to fix rotation
*/

type Tetronimo struct {
	shape     [][]int
	height    int
	longitude int
}

type Well struct {
	debris_map [][]int
}

type Stats struct {
	t_count int // tetronimo count
	b_count int // block count
	score   int
	rows    int
	t_types []int
}

func main() {

	// init termbox
	err := tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	// define tetromino
	t_size := 4
	tetronimo := new(Tetronimo)
	set_tetronimo(tetronimo, t_size)

	// define well
	well_depth := 20
	well_width := 10
	well := new(Well)
	set_well(well, well_depth, well_width)
	draw_border(well)

	// define stats
	set_count := 7 // number of tetronimos in our set
	stats := new(Stats)
	set_stats(stats, set_count)

	// starting block
	_ = new_block(well, tetronimo, stats)

	// keyboard channel
	ck := make(chan rune)
	go keys_in(ck)

	// timer channel
	starting_speed := 1
	ct := make(chan rune)
	go t_timer(ct, starting_speed)

	// main game loop
	for keep_going := true; keep_going == true; {

		// get input from keyboard or timer
		var somechar rune
		select {
		case somechar = <-ct:
			go t_timer(ct, get_speed(stats))
		case somechar = <-ck:
			go keys_in(ck)
		}

		// act on input
		operation := get_key(somechar)

		// quit
		if operation == "quit" {
			break
		}

		// pause
		if operation == "pause" {
			_ = tb.PollEvent().Ch
		}

		// attempt to move block
		block_status := block_action(well, tetronimo, operation)

		// new block
		if block_status == "stuck" {

			create_debris(well, tetronimo)
			clear_debris(well, stats)
			draw_debris(well)

			// game over if new block collides with debris
			blocked := new_block(well, tetronimo, stats)
			if blocked == true {
				keep_going = false
			}

		}

		show_stats(stats)
		tb.Flush()

	}

	tb.Close()
	end_stats(stats)
	fmt.Print("Game over\n")
	os.Exit(0)

}

func debug_stats(height int, show_text string, show_val int) {

	bh_status := fmt.Sprintf("%s : %d     ", show_text, show_val)
	print_tb(0, height, 0, 0, bh_status)
	_ = tb.PollEvent().Ch

}

func show_stats(stats *Stats) {

	print_tb(0, 4, 0, 0, fmt.Sprintf("tetronimos : %d     ", stats.t_count))
	print_tb(0, 5, 0, 0, fmt.Sprintf("blocks     : %d     ", stats.b_count))
	print_tb(0, 6, 0, 0, fmt.Sprintf("rows       : %d     ", stats.rows))
	print_tb(0, 7, 0, 0, fmt.Sprintf("score      : %d     ", stats.score))
	print_tb(0, 8, 0, 0, fmt.Sprintf("speed      : %d     ", get_speed(stats)))

	print_tb(0, 10, 0, 0, fmt.Sprintf("tet O      : %d     ", stats.t_types[0]))
	print_tb(0, 11, 0, 0, fmt.Sprintf("tet T      : %d     ", stats.t_types[1]))
	print_tb(0, 12, 0, 0, fmt.Sprintf("tet L      : %d     ", stats.t_types[2]))
	print_tb(0, 13, 0, 0, fmt.Sprintf("tet J      : %d     ", stats.t_types[3]))
	print_tb(0, 14, 0, 0, fmt.Sprintf("tet S      : %d     ", stats.t_types[4]))
	print_tb(0, 15, 0, 0, fmt.Sprintf("tet Z      : %d     ", stats.t_types[5]))
	print_tb(0, 16, 0, 0, fmt.Sprintf("tet I      : %d     ", stats.t_types[6]))

}

func end_stats(stats *Stats) {

	fmt.Printf("tetronimos : %d\n", stats.t_count)
	fmt.Printf("blocks     : %d\n", stats.b_count)
	fmt.Printf("rows       : %d\n", stats.rows)
	fmt.Printf("score      : %d\n", stats.score)
	fmt.Printf("speed      : %d\n", get_speed(stats))

	fmt.Printf("tet O      : %d\n", stats.t_types[0])
	fmt.Printf("tet T      : %d\n", stats.t_types[1])
	fmt.Printf("tet L      : %d\n", stats.t_types[2])
	fmt.Printf("tet J      : %d\n", stats.t_types[3])
	fmt.Printf("tet S      : %d\n", stats.t_types[4])
	fmt.Printf("tet Z      : %d\n", stats.t_types[5])
	fmt.Printf("tet I      : %d\n", stats.t_types[6])

}

func check_collisions(well *Well, this_tetronimo *Tetronimo, operation string) bool {

	if operation == "harddrop" {
		return false
	}

	ghost_tetronimo := new(Tetronimo)
	set_tetronimo(ghost_tetronimo, len(this_tetronimo.shape))
	clone_tetronimo(this_tetronimo, ghost_tetronimo)

	move_block(ghost_tetronimo, well, operation)

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

func move_block(tetronimo *Tetronimo, well *Well, operation string) {
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
		sound_depth(tetronimo, well)
	}
}

func sound_depth(this_tetronimo *Tetronimo, well *Well) {

	ghost_tetronimo := new(Tetronimo)
	set_tetronimo(ghost_tetronimo, len(this_tetronimo.shape))
	clone_tetronimo(this_tetronimo, ghost_tetronimo)

	for ghost_height := ghost_tetronimo.height; ghost_height >= 0; ghost_height-- {
		ghost_tetronimo.height = ghost_height
		blocked := check_collisions(well, ghost_tetronimo, "dropone")
		if blocked == true {
			this_tetronimo.height = ghost_height
			return
		}
	}

}

func draw_tetronimo(well *Well, operation string, this_tetronimo *Tetronimo) {

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
					color = set_color(this_tetronimo.shape[t_vert][t_horz])
				}
				height := (well_bottom - this_tetronimo.height + t_vert)
				longitude := (well_left + (this_tetronimo.longitude * 2) + (t_horz * 2))
				tb.SetCell(longitude, height, 0, 0, color)
				tb.SetCell(longitude+1, height, 0, 0, color)
			}
		}
	}

}

func draw_border(well *Well) {

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

func new_block(well *Well, tetronimo *Tetronimo, stats *Stats) bool {

	tetronimo.height = len(well.debris_map) - 1
	tetronimo.longitude = len(well.debris_map[0]) / 2

	rand_block(tetronimo, stats)

	for t_vert := range tetronimo.shape {
		for t_horz := range tetronimo.shape[t_vert] {
			t_bit_vert := tetronimo.height - t_vert
			t_bit_horz := tetronimo.longitude + t_horz
			if tetronimo.shape[t_vert][t_horz] > 0 {
				if well.debris_map[t_bit_vert][t_bit_horz] > 0 {
					return true // game over!
				}
			}
		}
	}

	return false

}

func draw_debris(well *Well) {

	term_col, term_row := tb.Size()
	well_depth := len(well.debris_map)
	vert_headroom := int((term_row-well_depth)/2) - 1

	for row := range well.debris_map {
		for col := range well.debris_map[row] {
			row_loc := vert_headroom + len(well.debris_map) - row
			col_loc := ((term_col / 2) - len(well.debris_map[0])) + (col * 2)

			color := set_color(well.debris_map[row][col])
			tb.SetCell(col_loc, row_loc, 0, 0, color)
			tb.SetCell(col_loc+1, row_loc, 0, 0, color)
		}
	}

}

func clear_debris(well *Well, stats *Stats) {

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

	stats.rows += delete_rows

	new_debris := make([][]int, len(well.debris_map))
	new_rows := 0
	for d_vert := range well.debris_map {
		if clear_rows[d_vert] != 1 {
			new_debris[new_rows] = well.debris_map[d_vert]
			new_rows++
		}
	}

	for i := (len(well.debris_map) - delete_rows); i < len(well.debris_map); i++ {
		fresh_row := make([]int, well_width)
		new_debris[i] = fresh_row
	}

	for this_row := range new_debris {
		well.debris_map[this_row] = new_debris[this_row]
	}

}

func rand_block(this_tetronimo *Tetronimo, stats *Stats) {

	set_count := 7

	tetronimo_set := make([][][]int, set_count+1)
	for set_num := 0; set_num <= set_count; set_num++ {

		tetro_def := make([][]int, 2)
		tetronimo_set[set_num] = tetro_def

		tetro_row := make([][]int, len(this_tetronimo.shape))
		for i := 0; i < len(this_tetronimo.shape); i++ {
			tetro_col := make([]int, len(this_tetronimo.shape))
			tetro_row[i] = tetro_col
		}
		tetronimo_set[set_num] = tetro_row
	}

	// there is no tetronimo_set[0]

	// define "O" block
	tetronimo_set[1][0][0] = 1
	tetronimo_set[1][0][1] = 1
	tetronimo_set[1][1][0] = 1
	tetronimo_set[1][1][1] = 1

	// define "T" block
	tetronimo_set[2][0][0] = 2
	tetronimo_set[2][0][1] = 2
	tetronimo_set[2][0][2] = 2
	tetronimo_set[2][1][1] = 2

	// define "L" block
	tetronimo_set[3][0][0] = 3
	tetronimo_set[3][1][0] = 3
	tetronimo_set[3][2][0] = 3
	tetronimo_set[3][2][1] = 3

	// define "J" block
	tetronimo_set[4][0][1] = 4
	tetronimo_set[4][1][1] = 4
	tetronimo_set[4][2][0] = 4
	tetronimo_set[4][2][1] = 4

	// define "S" block
	tetronimo_set[5][0][0] = 5
	tetronimo_set[5][1][0] = 5
	tetronimo_set[5][1][1] = 5
	tetronimo_set[5][2][1] = 5

	// define "Z" block
	tetronimo_set[6][0][1] = 6
	tetronimo_set[6][1][0] = 6
	tetronimo_set[6][1][1] = 6
	tetronimo_set[6][2][0] = 6

	// define "I" block
	tetronimo_set[7][0][1] = 7
	tetronimo_set[7][1][1] = 7
	tetronimo_set[7][2][1] = 7
	tetronimo_set[7][3][1] = 7
	tetronimo_set[7][1][0] = -1 // phantom blocks to help rotation
	tetronimo_set[7][2][2] = -1

	rand.Seed(time.Now().Unix())
	rand_tetro := rand.Intn(set_count)

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

	stats.t_count += 1
	stats.b_count += b_count
	stats.t_types[rand_tetro] += 1

}

func block_action(well *Well, tetronimo *Tetronimo, operation string) string {

	block_status := "free"

	blocked := check_collisions(well, tetronimo, operation)
	if blocked == true {
		if operation == "dropone" {
			block_status = "stuck"
		}
		return block_status
	}

	draw_tetronimo(well, "erase", tetronimo)

	move_block(tetronimo, well, operation)

	if operation == "harddrop" {
		block_status = "stuck"
	}

	draw_tetronimo(well, "draw", tetronimo)

	return block_status

}

func rotate_tetronimo(this_tetronimo *Tetronimo) {

	// rotate
	ghost_tetronimo := new(Tetronimo)
	set_tetronimo(ghost_tetronimo, len(this_tetronimo.shape))
	ghost_tetronimo.height = this_tetronimo.height
	ghost_tetronimo.longitude = this_tetronimo.longitude
	for row := range ghost_tetronimo.shape {
		rotated_col := (len(ghost_tetronimo.shape) - 1) - row
		for col := range ghost_tetronimo.shape[row] {
			rotated_row := col
			ghost_tetronimo.shape[row][col] = this_tetronimo.shape[rotated_row][rotated_col]
		}
	}

	top_left(ghost_tetronimo)

	clone_tetronimo(ghost_tetronimo, this_tetronimo)

}

func create_debris(well *Well, tetronimo *Tetronimo) {

	for t_vert := range tetronimo.shape {
		for t_horz := range tetronimo.shape[t_vert] {
			t_bit_vert := tetronimo.height - t_vert
			t_bit_horz := tetronimo.longitude + t_horz
			if tetronimo.shape[t_vert][t_horz] > 0 {
				well.debris_map[t_bit_vert][t_bit_horz] = tetronimo.shape[t_vert][t_horz]
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

func clone_tetronimo(orig_tetronimo, new_tetronimo *Tetronimo) {

	new_tetronimo.height = orig_tetronimo.height
	new_tetronimo.longitude = orig_tetronimo.longitude

	for row := 0; row < len(orig_tetronimo.shape); row++ {
		for col := 0; col < len(orig_tetronimo.shape[0]); col++ {
			new_tetronimo.shape[row][col] = orig_tetronimo.shape[row][col]
		}
	}

}

func set_tetronimo(tetronimo *Tetronimo, t_size int) {

	this_tetronimo := make([][]int, t_size)
	for i := 0; i < t_size; i++ {
		tetro_row := make([]int, t_size)
		this_tetronimo[i] = tetro_row
	}

	tetronimo.height = 3
	tetronimo.longitude = 4
	tetronimo.shape = this_tetronimo

}

func set_well(well *Well, well_depth, well_width int) {

	this_well := make([][]int, well_depth)
	for i := 0; i < well_depth; i++ {
		this_row := make([]int, well_width)
		this_well[i] = this_row
	}

	well.debris_map = this_well

}

func set_stats(stats *Stats, set_count int) {

	t_types := make([]int, set_count)

	stats.t_types = t_types

}

func get_key(somechar rune) string {

	switch {
	case somechar == 113: // q
		return "quit"
	case somechar == 106: // j
		return "left"
	case somechar == 108: // l
		return "right"
	case somechar == 110: // n
		return "dropone"
	case somechar == 112: // p
		return "pause"
	case somechar == 107: // k
		return "rotate"
	case somechar == 0: // [space]
		return "harddrop"
	}
	return "hold" // do nothing
}

func get_speed(stats *Stats) int {
	speed := 1
	if stats.rows > 10 {
		speed = int(stats.rows / 10)
	}
	return speed
}

func top_left(this_tetronimo *Tetronimo) {

	row_top := 0
	col_left := 0
	row_offset := 0
	col_offset := 0
	for row := range this_tetronimo.shape {
		for _, col_val := range this_tetronimo.shape[row] {
			row_top += int(col_val)
		}
		if row_top == 0 {
			row_offset += 1
		}
		col_left += int(this_tetronimo.shape[row][0])
	}

	if col_left == 0 {
		col_offset = 1
	}

	if row_offset > 2 {
		row_offset = 2
	}

	ghost_tetronimo := new(Tetronimo)
	set_tetronimo(ghost_tetronimo, len(this_tetronimo.shape))
	ghost_tetronimo.height = this_tetronimo.height
	ghost_tetronimo.longitude = this_tetronimo.longitude

	for row := range this_tetronimo.shape {
		this_row := row - row_offset
		if this_row < 0 {
			for col := range this_tetronimo.shape[row] {
				ghost_tetronimo.shape[len(this_tetronimo.shape)+this_row][col] = 0
			}
		} else {
			for col := range this_tetronimo.shape[this_row] {
				this_col := col - col_offset
				if this_col < 0 {
					ghost_tetronimo.shape[this_row][len(this_tetronimo.shape[row])-col_offset] = 0
				} else {
					ghost_tetronimo.shape[this_row][this_col] = this_tetronimo.shape[row][col]
				}
			}
		}
	}

	clone_tetronimo(ghost_tetronimo, this_tetronimo)

}

func set_color(color int) tb.Attribute {

	var colorname tb.Attribute

	switch {
	case color == -1:
		colorname = tb.ColorDefault
	case color == 0:
		colorname = tb.ColorDefault
	case color == 1:
		colorname = tb.ColorBlue
	case color == 2:
		colorname = tb.ColorYellow
	case color == 3:
		colorname = tb.ColorMagenta
	case color == 4:
		colorname = tb.ColorWhite
	case color == 5:
		colorname = tb.ColorGreen
	case color == 6:
		colorname = tb.ColorCyan
	case color == 7:
		colorname = tb.ColorRed
	}

	return colorname

}
