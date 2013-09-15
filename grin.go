package main

// http://tetrisconcept.net/wiki/Tetris_Guideline

import "fmt"
import "math/rand"
import "time"
import "os"
import "os/user"
import "path"
import "flag"
import "database/sql"

import _ "github.com/mattn/go-sqlite3"
import tb "github.com/nsf/termbox-go"
import pieces "github.com/nickaubert/grin/pieces"

/*
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
		High scores in sqlite?
		Blocks should push away when rotating if blocked on right
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
	p_types   map[string]int
	piece_set map[string]bool
}

func main() {

	// some defaults
	default_width := 10
	default_depth := 20

	// get arguments
	well_width := flag.Int("w", default_width, "well width")
	well_depth := flag.Int("d", default_depth, "well depth")
	use_extd := flag.Bool("e", false, "Use extended pieces set")
	junk_level := flag.Int("j", 0, "Starting junk")
	flag.Parse()

	// check db file
	db := init_db("/home/nick/.grin/score.db")
	defer db.Close()

	// init termbox
	err := tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	// define stats
	stats := new(Stats)
	set_count := 30 // need to figure this dynamically
	max_stats := 10 // top x score
	set_stats(stats, set_count)
	stats.piece_set["extd"] = *use_extd

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
			if operation == "pause" {
				_ = tb.PollEvent().Ch
			}
			go keys_in(ck)
		}

		// quit
		if operation == "quit" {
			break
		}

		// attempt to move block get status
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
	fmt.Print("Game over\n")
	end_stats(stats, well, max_stats, db)
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

	print_tb(0, vert_headroom+0, 0, 0, fmt.Sprintf("score  : %d", stats.score))
	print_tb(0, vert_headroom+1, 0, 0, fmt.Sprintf("rows   : %d", stats.rows))
	print_tb(0, vert_headroom+2, 0, 0, fmt.Sprintf("speed  : %d", get_speed(stats)))
	print_tb(0, vert_headroom+3, 0, 0, fmt.Sprintf("pieces : %d", stats.p_count))
	print_tb(0, vert_headroom+4, 0, 0, fmt.Sprintf("blocks : %d", stats.b_count))

	ShowSet := pieces.SetBasic()
	if stats.piece_set["extd"] == true {
		ExtendedSet := pieces.SetExtended()
		for t_num := range ExtendedSet {
			ShowSet = append(ShowSet, ExtendedSet[t_num])
		}
	}

	p_row := vert_headroom + 6
	for _, value := range ShowSet {
		if p_row >= term_row {
			break
		}
		if stats.p_types[value.Name] > 0 {
			print_tb(0, p_row, 0, 0, fmt.Sprintf("%s  : %d", value.Name, stats.p_types[value.Name]))
			p_row += 1
		}
	}

}

func end_stats(stats *Stats, well *Well, max_stats int, db *sql.DB) {

	fmt.Printf("score     : %d\n", stats.score)
	fmt.Printf("rows      : %d\n", stats.rows)
	fmt.Printf("top speed : %d\n", get_speed(stats))
	fmt.Printf("pieces    : %d\n", stats.p_count)
	fmt.Print("")

	ShowSet := pieces.SetBasic()
	if stats.piece_set["extd"] == true {
		ShowSet = pieces.SetExtended()
	}

	for _, value := range ShowSet {
		fmt.Printf("%s    : %d\n", value.Name, stats.p_types[value.Name])
	}

	timenow := update_db(stats, well, max_stats, db)
	show_db_scores(stats, well, db, timenow)

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
					return "blocked_right"
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

	rand.Seed(time.Now().Unix())

	BasicSet := pieces.SetBasic()
	ExtendedSet := pieces.SetExtended()

	FullSet := BasicSet
	for t_num := range ExtendedSet {
		FullSet = append(FullSet, ExtendedSet[t_num])
	}

	// this logic favors pieces from basic set
	ChosenSet := BasicSet
	if stats.piece_set["extd"] == true {
		rand_set := rand.Intn(2)
		if rand_set == 1 {
			ChosenSet = FullSet
		}
	}

	rand_piece := rand.Intn(len(ChosenSet))

	b_count := 4 // assume always tetro for now
	chosen_piece := ChosenSet[rand_piece]
	copy_shape(chosen_piece.Shape, this_piece.shape)

	piece_name := chosen_piece.Name

	stats.p_count += 1
	stats.b_count += b_count
	stats.t_types[rand_piece] += 1
	stats.p_types[piece_name] += 1

}

func block_action(well *Well, piece *Tetronimo, operation string) string {

	block_status := "free"

	if operation == "pause" {
		return block_status
	}

	piece_status := check_collisions(well, piece, operation)

	// if piece is blocked on right wall, try shifting it to the left
	left_shift := 0
	if piece_status == "blocked_right" {

		if operation != "rotate" {
			return "blocked"
		}

		ghost_piece := new(Tetronimo)
		set_piece(ghost_piece, len(piece.shape))
		clone_piece(piece, ghost_piece)
		ghost_piece.longitude -= 1
		ghost_status := check_collisions(well, ghost_piece, operation)

		if ghost_status == "blocked" {
			return "blocked"
		}

		if ghost_status == "blocked_right" {
			return "blocked"
		}

		// shift to the left
		left_shift += 1

	}

	if piece_status == "blocked" {
		if operation == "dropone" {
			block_status = "stuck"
		}
		return block_status
	}

	draw_piece(well, "erase", piece)

	piece.longitude -= left_shift

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

	piece_names := make(map[string]int)
	stats.p_types = piece_names

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

func init_db(filepath string) *sql.DB {

	// fmt.Printf("no file %s\n", filepath)
	db_dir := path.Dir(filepath)
	init_dir(db_dir)

	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, file_err := os.Stat(filepath)
	if file_err == nil {
		return db
	}

	create_table_sql := `create table stats (
		score integer not null, 
		timestamp integer, 
		user text,
		width integer not null default 10,
		depth integer not null default 20,
		extended_set integer not null default 0
	)`
	_, err = db.Exec(create_table_sql)
	if err != nil {
		fmt.Printf("ERROR initiating %s : %q:\n", filepath, err)
		os.Exit(1)
	}

	return db

}

func init_dir(db_dir string) {
	db_info, db_err := os.Stat(db_dir)
	if db_err != nil {
		mkdb_err := os.Mkdir(db_dir, 0755)
		if mkdb_err != nil {
			fmt.Println(mkdb_err)
			os.Exit(1)
		}
		return
	}
	if db_info.IsDir() == false {
		fmt.Printf("ERROR: %s exists but is not a directory\n", db_dir)
		os.Exit(1)
	}
}

func update_db(stats *Stats, well *Well, max_stats int, db *sql.DB) int64 {

	well_depth := len(well.debris_map)
	well_width := len(well.debris_map[0])
	extended_set := 0
	if stats.piece_set["extd"] == true {
		extended_set = 1
	}

	check_score_sql := fmt.Sprintf(`
		select 
			score 
		from 
			stats 
		where
			width = %d
		and
			depth = %d
		and
			extended_set = %d
		order by 
			score 
		limit 1
	`, well_width, well_depth, extended_set)

	rows, err := db.Query(check_score_sql)
	if err != nil {
		fmt.Println("ERROR querying db\n")
		fmt.Println(check_score_sql)
		os.Exit(1)
	}
	defer rows.Close()

	var min_score int
	for rows.Next() {
		rows.Scan(&min_score)
	}
	rows.Close()

	row_count_sql := fmt.Sprintf(`
		select 
			count(*)
		from 
			stats
		where
			width = %d
		and
			depth = %d
		and
			extended_set = %d
	`, well_width, well_depth, extended_set)
	rows, err = db.Query(row_count_sql)
	if err != nil {
		fmt.Println("ERROR querying db\n")
		fmt.Println(row_count_sql)
		os.Exit(1)
	}
	defer rows.Close()

	var record_count int
	for rows.Next() {
		rows.Scan(&record_count)
	}
	rows.Close()

	timenow := time.Now().Unix()
	username := get_playername()

	if min_score > stats.score {
		if record_count >= max_stats {
			return timenow
		}
	}

	if record_count >= max_stats {
		remove_old_sql := fmt.Sprintf(`
			delete from 
				stats 
			where 
				score <= %d
			and 
				width = %d
			and
				depth = %d
			and
				extended_set = %d
		`, min_score, well_width, well_depth, extended_set)
		_, err = db.Exec(remove_old_sql)
		if err != nil {
			fmt.Printf("ERROR updating sql ==%s== : %q:\n", remove_old_sql, err)
			os.Exit(1)
		}
	}

	update_score_sql := fmt.Sprintf(
		`insert into stats(
			score, 
				timestamp, 
			user,
			width,
			depth,
			extended_set
		) values (
			%d, 
			%d, 
			'%s',
			%d, 
			%d, 
			%d 
		)`, stats.score, timenow, username, well_width, well_depth, extended_set)
	_, err = db.Exec(update_score_sql)
	if err != nil {
		fmt.Printf("ERROR updating sql ==%s== : %q:\n", update_score_sql, err)
		os.Exit(1)
	}

	return timenow

}

func get_playername() string {

	thisuser, err := user.Current()
	if err != nil {
		return "null"
	}

	return thisuser.Username

}

func show_db_scores(stats *Stats, well *Well, db *sql.DB, timenow int64) {

	well_depth := len(well.debris_map)
	well_width := len(well.debris_map[0])
	extended_set := 0
	if stats.piece_set["extd"] == true {
		extended_set = 1
	}

	get_scores_sql := fmt.Sprintf(`
		select 
			score, 
			timestamp, 
			user
		from 
			stats
		where
			width = %d
		and
			depth = %d
		and
			extended_set = %d
		order by 
			score 
		desc
	`, well_width, well_depth, extended_set)

	scores, err := db.Query(get_scores_sql)
	if err != nil {
		fmt.Println("ERROR querying db\n")
		os.Exit(1)
	}
	defer scores.Close()

	var show_scores [][]string
	max_score_len := len("score")
	max_player_len := len("player")

	fmt.Print("\nHigh Scores:\n   score - player - date\n")
	for scores.Next() {
		var score int
		var timestamp int64
		var player string
		scores.Scan(&score, &timestamp, &player)

		score_len := len(fmt.Sprintf("%d", score))
		if score_len > max_score_len {
			max_score_len = score_len
		}

		player_len := len(player)
		if player_len > max_player_len {
			max_player_len = player_len
		}

		asterisk := "  "
		if timestamp == timenow {
			asterisk = " *"
		}

		time := time.Unix(int64(timestamp), 0)
		var show_score []string
		show_score = append(show_score, fmt.Sprintf("%d", score))
		show_score = append(show_score, player)
		show_score = append(show_score, fmt.Sprintf("%04d-%02d-%02d", time.Year(), time.Month(), time.Day()))
		show_score = append(show_score, asterisk)
		show_scores = append(show_scores, show_score)
	}
	scores.Close()

	for this_score := range show_scores {
		format := fmt.Sprintf("%%s %%%ds - %%%ds - %%s\n", max_score_len, max_player_len)
		fmt.Printf(format, show_scores[this_score][3], show_scores[this_score][0], show_scores[this_score][1], show_scores[this_score][2])
	}

}
