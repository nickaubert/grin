package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "fmt"
import "math/rand"
import "time"
import "os"
import "flag"

import tb "github.com/nsf/termbox-go"
import pieces "github.com/nickaubert/grin/pieces"

/*
	Improvements:
		High scores in sqlite?
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
				( 1 point x speed ) per level for each distance dropped
				0 points for clearing rows
		Print score, stats on exit
		*** CONVERT FROM goncurses to termbox
		Cleanup functions, data objects (can always refactor)
		Phantom blocks to fix rotation
		Arrow key control
		Adjustible well size and tetronimo set
		Random debris map at start
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
	p_count   int // piece count
	b_count   int // block count
	score     int
	rows      int
	t_types   []int
	piece_set map[string]bool
}

func main() {

	// some defaults
	default_width := 10
	default_depth := 20

	// get arguments
	well_width := flag.Int("w", default_width, "well width")
	well_depth := flag.Int("d", default_depth, "well depth")
	use_huge := flag.Bool("u", false, "Use huge pieces")
	use_pento := flag.Bool("p", false, "Use pentomino pieces")
	use_tiny := flag.Bool("t", false, "Use tiny pieces")
	junk_level := flag.Int("j", 0, "Starting junk")
	flag.Parse()

	// init termbox
	err := tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	// define stats
	stats := new(Stats)
	set_count := 30 // need to figure this dynamically
	set_stats(stats, set_count)
	stats.piece_set["basic"] = true
	stats.piece_set["huge"] = *use_huge
	stats.piece_set["pento"] = *use_pento
	stats.piece_set["tiny"] = *use_tiny

	// define piece set
	p_size := 4
	piece := new(Tetronimo)
	set_piece(piece, p_size)

	// define well
	well := new(Well)
	set_well(well, *well_depth, *well_width, *junk_level)

	// minimum window size before drawing borders
	check_size(well, piece, *junk_level)
	draw_border(well)

	// draw starting junk
	draw_debris(well)

	// starting block
	_ = new_piece(well, piece, stats)

	// keyboard channel
	ck := make(chan string)
	go keys_in(ck)

	// timer channel
	starting_speed := 1
	ct := make(chan string)
	go t_timer(ct, starting_speed)

	// main game loop
	for keep_going := true; keep_going == true; {

		// get input from keyboard or timer
		var operation string
		select {
		case operation = <-ct:
			go t_timer(ct, get_speed(stats))
		case operation = <-ck:
			go keys_in(ck)
		}

		// quit
		if operation == "quit" {
			break
		}

		// pause
		if operation == "pause" {
			_ = tb.PollEvent().Ch
		}

		// attempt to move block
		last_height := piece.height
		block_status := block_action(well, piece, operation)

		// new block
		if block_status == "stuck" {

			create_debris(well, piece)
			clear_debris(well, stats)
			draw_debris(well)

			points := last_height - piece.height
			if points == 0 {
				points = 1
			}
			stats.score += get_speed(stats) * points

			// game over if new block collides with debris
			blocked := new_piece(well, piece, stats)
			if blocked == true {
				keep_going = false
			}

		}

		show_stats(stats, well)
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

func show_stats(stats *Stats, well *Well) {

	well_depth := len(well.debris_map)
	_, term_row := tb.Size()
	vert_headroom := int((term_row-well_depth)/2) - 1

	print_tb(0, vert_headroom+0, 0, 0, fmt.Sprintf("pieces : %d", stats.p_count))
	print_tb(0, vert_headroom+1, 0, 0, fmt.Sprintf("blocks : %d", stats.b_count))
	print_tb(0, vert_headroom+2, 0, 0, fmt.Sprintf("rows   : %d", stats.rows))
	print_tb(0, vert_headroom+3, 0, 0, fmt.Sprintf("score  : %d", stats.score))
	print_tb(0, vert_headroom+4, 0, 0, fmt.Sprintf("speed  : %d", get_speed(stats)))

	print_tb(0, vert_headroom+6, 0, 0, fmt.Sprintf("tet O  : %d", stats.t_types[0]))
	print_tb(0, vert_headroom+7, 0, 0, fmt.Sprintf("tet T  : %d", stats.t_types[1]))
	print_tb(0, vert_headroom+8, 0, 0, fmt.Sprintf("tet L  : %d", stats.t_types[2]))
	print_tb(0, vert_headroom+9, 0, 0, fmt.Sprintf("tet J  : %d", stats.t_types[3]))
	print_tb(0, vert_headroom+10, 0, 0, fmt.Sprintf("tet S  : %d", stats.t_types[4]))
	print_tb(0, vert_headroom+11, 0, 0, fmt.Sprintf("tet Z  : %d", stats.t_types[5]))
	print_tb(0, vert_headroom+12, 0, 0, fmt.Sprintf("tet I  : %d", stats.t_types[6]))

}

func end_stats(stats *Stats) {

	fmt.Printf("tetronimos : %d\n", stats.p_count)
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

func check_collisions(well *Well, this_piece *Tetronimo, operation string) string {

	if operation == "harddrop" {
		return "free"
	}

	ghost_piece := new(Tetronimo)
	set_piece(ghost_piece, len(this_piece.shape))
	clone_piece(this_piece, ghost_piece)

	move_piece(ghost_piece, well, operation)

	for t_vert := range ghost_piece.shape {
		for t_horz := range ghost_piece.shape[t_vert] {

			t_bit_vert := ghost_piece.height - t_vert
			t_bit_horz := ghost_piece.longitude + t_horz

			if ghost_piece.shape[t_vert][t_horz] > 0 {
				if t_bit_horz < 0 {
					return "blocked"
				}
				if t_bit_horz >= len(well.debris_map[0]) {
					return "blocked"
				}
				if t_bit_vert < 0 {
					return "blocked"
				}
				if well.debris_map[t_bit_vert][t_bit_horz] > 0 {
					return "blocked"
				}
			}
		}
	}

	return "free"
}

func move_piece(piece *Tetronimo, well *Well, operation string) {
	switch {
	case operation == "left":
		piece.longitude--
	case operation == "right":
		piece.longitude++
	case operation == "rotate":
		rotate_piece(piece)
	case operation == "dropone":
		piece.height--
	case operation == "harddrop":
		sound_depth(piece, well)
	}
}

func sound_depth(this_piece *Tetronimo, well *Well) {

	ghost_piece := new(Tetronimo)
	set_piece(ghost_piece, len(this_piece.shape))
	clone_piece(this_piece, ghost_piece)

	for ghost_height := ghost_piece.height; ghost_height >= 0; ghost_height-- {
		ghost_piece.height = ghost_height
		piece_status := check_collisions(well, ghost_piece, "dropone")
		if piece_status == "blocked" {
			this_piece.height = ghost_height
			return
		}
	}

}

func draw_piece(well *Well, operation string, this_piece *Tetronimo) {

	well_depth := len(well.debris_map)

	term_col, term_row := tb.Size()
	vert_headroom := int((term_row-well_depth)/2) - 1

	well_bottom := len(well.debris_map) + vert_headroom
	well_left := ((term_col / 2) - len(well.debris_map[0]))

	for t_vert := range this_piece.shape {
		for t_horz := range this_piece.shape[t_vert] {
			if this_piece.shape[t_vert][t_horz] > 0 {
				color := tb.ColorDefault
				if operation == "draw" {
					color = set_color(this_piece.shape[t_vert][t_horz])
				}
				height := (well_bottom - this_piece.height + t_vert)
				longitude := (well_left + (this_piece.longitude * 2) + (t_horz * 2))
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

func new_piece(well *Well, piece *Tetronimo, stats *Stats) bool {

	piece.height = len(well.debris_map) - 1
	piece.longitude = len(well.debris_map[0]) / 2

	rand_piece(piece, stats)

	for t_vert := range piece.shape {
		for t_horz := range piece.shape[t_vert] {
			t_bit_vert := piece.height - t_vert
			t_bit_horz := piece.longitude + t_horz
			if piece.shape[t_vert][t_horz] > 0 {
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

func rand_piece(this_piece *Tetronimo, stats *Stats) {

	var basic_set = [][][]int{
		pieces.BasicO,
		pieces.BasicT,
		pieces.BasicL,
		pieces.BasicJ,
		pieces.BasicS,
		pieces.BasicZ,
		pieces.BasicI,
	}

	var huge_set = [][][]int{
		pieces.HugeL,
		pieces.HugeJ,
		pieces.HugeU,
	}

	var pento_set = [][][]int{
		pieces.PentoF,
		pieces.PentoL,
		pieces.PentoN,
		pieces.PentoP,
		pieces.PentoT,
		pieces.PentoT,
		pieces.PentoU,
		pieces.PentoV,
		pieces.PentoW,
		pieces.PentoX,
		pieces.PentoY,
		pieces.PentoZ,
	}

	var tiny_set = [][][]int{
		pieces.TinyO,
		pieces.TinyI,
		pieces.SmallL,
		pieces.SmallI,
	}

	var full_set [][][]int

	if stats.piece_set["basic"] == true {
		for t_num := range basic_set {
			full_set = append(full_set, basic_set[t_num])
		}
	}

	if stats.piece_set["huge"] == true {
		for t_num := range huge_set {
			full_set = append(full_set, huge_set[t_num])
		}
	}

	if stats.piece_set["pento"] == true {
		for t_num := range pento_set {
			full_set = append(full_set, pento_set[t_num])
		}
	}

	if stats.piece_set["tiny"] == true {
		for t_num := range tiny_set {
			full_set = append(full_set, tiny_set[t_num])
		}
	}

	rand.Seed(time.Now().Unix())
	rand_piece := rand.Intn(len(full_set))

	b_count := 4 // assume always tetro for now
	copy_shape(full_set[rand_piece], this_piece.shape)

	stats.p_count += 1
	stats.b_count += b_count
	stats.t_types[rand_piece] += 1

}

func block_action(well *Well, piece *Tetronimo, operation string) string {

	block_status := "free"

	piece_status := check_collisions(well, piece, operation)
	if piece_status == "blocked" {
		if operation == "dropone" {
			block_status = "stuck"
		}
		return block_status
	}

	draw_piece(well, "erase", piece)

	move_piece(piece, well, operation)

	if operation == "harddrop" {
		block_status = "stuck"
	}

	draw_piece(well, "draw", piece)

	return block_status

}

func rotate_piece(this_piece *Tetronimo) {

	// rotate
	ghost_piece := new(Tetronimo)
	set_piece(ghost_piece, len(this_piece.shape))
	ghost_piece.height = this_piece.height
	ghost_piece.longitude = this_piece.longitude
	for row := range ghost_piece.shape {
		rotated_col := (len(ghost_piece.shape) - 1) - row
		for col := range ghost_piece.shape[row] {
			rotated_row := col
			ghost_piece.shape[row][col] = this_piece.shape[rotated_row][rotated_col]
		}
	}

	top_left(ghost_piece)

	clone_piece(ghost_piece, this_piece)

}

func create_debris(well *Well, piece *Tetronimo) {

	for t_vert := range piece.shape {
		for t_horz := range piece.shape[t_vert] {
			t_bit_vert := piece.height - t_vert
			t_bit_horz := piece.longitude + t_horz
			if piece.shape[t_vert][t_horz] > 0 {
				well.debris_map[t_bit_vert][t_bit_horz] = piece.shape[t_vert][t_horz]
			}
		}
	}

}

func keys_in(ck chan string) {

	key := tb.PollEvent()

	operation := get_key(key)
	ck <- operation

}

func t_timer(ct chan string, speed int) {
	mseconds := time.Duration(1000 / speed)
	time.Sleep(mseconds * time.Millisecond)
	ct <- "dropone"
}

func error_out(message string) {

	tb.Close()
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

func debug_tb(x, y int, fg, bg tb.Attribute, msg string) {
	for _, c := range msg {
		tb.SetCell(x, y, c, fg, bg)
		x++
	}
	tb.Flush()
	_ = tb.PollEvent().Ch
}

func clone_piece(orig_piece, new_piece *Tetronimo) {

	new_piece.height = orig_piece.height
	new_piece.longitude = orig_piece.longitude
	copy_shape(orig_piece.shape, new_piece.shape)

}

func copy_shape(orig_shape, new_shape [][]int) {
	for row := 0; row < len(orig_shape); row++ {
		for col := 0; col < len(orig_shape[0]); col++ {
			new_shape[row][col] = orig_shape[row][col]
		}
	}
}

func set_piece(piece *Tetronimo, p_size int) {

	this_piece := make([][]int, p_size)
	for i := 0; i < p_size; i++ {
		piece_row := make([]int, p_size)
		this_piece[i] = piece_row
	}

	piece.height = 3
	piece.longitude = 4
	piece.shape = this_piece

}

func set_well(well *Well, well_depth, well_width, junk_level int) {

	well_debris := make([][]int, well_depth)
	rand.Seed(time.Now().Unix())
	for i := 0; i < well_depth; i++ {
		this_row := make([]int, well_width)
		// create some junk
		if i < junk_level {
			for j := 0; j < well_width; j++ {
				on_off := rand.Intn(2)
				if on_off > 0 {
					color := rand.Intn(6) + 1
					this_row[j] = color
				}
			}
			// ensure that at least one block in row is empty
			this_row[rand.Intn(well_width)] = 0
		}
		well_debris[i] = this_row
	}

	well.debris_map = well_debris

}

func set_stats(stats *Stats, set_count int) {

	t_types := make([]int, set_count)
	stats.t_types = t_types

	piece_set := make(map[string]bool)
	stats.piece_set = piece_set

}

func get_key(somekey tb.Event) string {

	switch {
	case somekey.Ch == 113: // q
		return "quit"
	case somekey.Ch == 106: // j
		return "left"
	case somekey.Ch == 108: // l
		return "right"
	case somekey.Ch == 110: // n
		return "dropone"
	case somekey.Ch == 112: // p
		return "pause"
	case somekey.Ch == 107: // k
		return "rotate"
	case somekey.Key == tb.KeyArrowUp:
		return "rotate"
	case somekey.Key == tb.KeyArrowLeft:
		return "left"
	case somekey.Key == tb.KeyArrowRight:
		return "right"
	case somekey.Key == tb.KeyArrowDown:
		return "dropone"
	case somekey.Key == tb.KeyPgup:
		return "pause"
	case somekey.Key == tb.KeyPgdn:
		return "harddrop"
	case somekey.Ch == 0: // [space]
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

func top_left(this_piece *Tetronimo) {

	row_top := 0
	col_left := 0
	row_offset := 0
	col_offset := 0
	for row := range this_piece.shape {
		for _, col_val := range this_piece.shape[row] {
			row_top += int(col_val)
		}
		if row_top == 0 {
			row_offset += 1
		}
		col_left += int(this_piece.shape[row][0])
	}

	if col_left == 0 {
		col_offset = 1
	}

	if row_offset > 2 {
		row_offset = 2
	}

	ghost_piece := new(Tetronimo)
	set_piece(ghost_piece, len(this_piece.shape))
	ghost_piece.height = this_piece.height
	ghost_piece.longitude = this_piece.longitude

	for row := range this_piece.shape {
		this_row := row - row_offset
		if this_row < 0 {
			for col := range this_piece.shape[row] {
				ghost_piece.shape[len(this_piece.shape)+this_row][col] = 0
			}
		} else {
			for col := range this_piece.shape[this_row] {
				this_col := col - col_offset
				if this_col < 0 {
					ghost_piece.shape[this_row][len(this_piece.shape[row])-col_offset] = 0
				} else {
					ghost_piece.shape[this_row][this_col] = this_piece.shape[row][col]
				}
			}
		}
	}

	clone_piece(ghost_piece, this_piece)

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

func set_key(this_key tb.Event) rune {

	var this_char rune
	switch {
	case this_key.Key == tb.KeyArrowLeft:
		this_char = 106 // j
	}

	return this_char

}

func check_size(well *Well, piece *Tetronimo, junk_level int) {

	if len(well.debris_map) < (junk_level + len(piece.shape)) {
		error_string := fmt.Sprintf("Well is too small for that much junk\n")
		error_out(error_string)
	}

	if len(well.debris_map) < len(piece.shape) {
		error_string := fmt.Sprintf("Well is too small\n")
		error_out(error_string)
	}

	if len(well.debris_map[0]) < len(piece.shape[0]) {
		error_string := fmt.Sprintf("Well is too small\n")
		error_out(error_string)
	}

	term_width, term_height := tb.Size()

	min_width := 15 + 2*len(well.debris_map[0])
	if term_width < min_width {
		error_string := fmt.Sprintf("Terminal window minimum width %d\n", min_width)
		error_out(error_string)
	}

	min_height := 13
	if min_height < len(well.debris_map)+1 {
		min_height = len(well.debris_map) + 1
	}

	if term_height < min_height {
		error_string := fmt.Sprintf("Terminal window minimum height %d\n", min_height)
		error_out(error_string)
	}

}
