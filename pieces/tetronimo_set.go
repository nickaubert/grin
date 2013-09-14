package tetronimo_set

type Piece struct {
	Shape [][]int
	Name  string
}

func SetBasic() []Piece {

	BasicPieces := make([]Piece, 7)

	BasicPieces[0].Name = "BasicO"
	BasicPieces[0].Shape = [][]int{
		{1, 1, 0, 0},
		{1, 1, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	BasicPieces[1].Name = "BasicT"
	BasicPieces[1].Shape = [][]int{
		{2, 2, 2, 0},
		{0, 2, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	BasicPieces[2].Name = "BasicL"
	BasicPieces[2].Shape = [][]int{
		{3, 0, 0, 0},
		{3, 0, 0, 0},
		{3, 3, 0, 0},
		{0, 0, 0, 0},
	}

	BasicPieces[3].Name = "BasicJ"
	BasicPieces[3].Shape = [][]int{
		{0, 4, 0, 0},
		{0, 4, 0, 0},
		{4, 4, 0, 0},
		{0, 0, 0, 0},
	}

	BasicPieces[4].Name = "BasicS"
	BasicPieces[4].Shape = [][]int{
		{0, 5, 5, 0},
		{5, 5, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	BasicPieces[5].Name = "BasicZ"
	BasicPieces[5].Shape = [][]int{
		{6, 6, 0, 0},
		{0, 6, 6, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	// includes negative phantom blocks to help rotation
	BasicPieces[6].Name = "BasicI"
	BasicPieces[6].Shape = [][]int{
		{0, 7, 0, 0},
		{-1, 7, 0, 0},
		{0, 7, -1, 0},
		{0, 7, 0, 0},
	}

	return BasicPieces

}

var HugeL = [][]int{
	{3, 0, 0, 0},
	{3, 0, 0, 0},
	{3, 0, 0, 0},
	{3, 3, 3, 0},
}

var HugeJ = [][]int{
	{0, 0, 0, 4},
	{0, 0, 0, 4},
	{0, 0, 0, 4},
	{0, 4, 4, 4},
}

var HugeU = [][]int{
	{2, 0, 0, 2},
	{2, 0, 0, 2},
	{2, 0, 0, 2},
	{2, 2, 2, 2},
}

func SetExtended() []Piece {

	ExtendedPieces := make([]Piece, 19)

	ExtendedPieces[0].Name = "PentoF"
	ExtendedPieces[0].Shape = [][]int{
		{0, 4, 4, 0},
		{4, 4, 0, 0},
		{0, 4, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[1].Name = "PentoFb"
	ExtendedPieces[1].Shape = [][]int{
		{2, 2, 0, 0},
		{0, 2, 2, 0},
		{0, 2, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[2].Name = "PentoL"
	ExtendedPieces[2].Shape = [][]int{
		{3, 0, 0, 0},
		{3, 0, 0, 0},
		{3, 0, 0, 0},
		{3, 3, 0, 0},
	}

	ExtendedPieces[0].Name = "PentoLb"
	ExtendedPieces[0].Shape = [][]int{
		{0, 4, 0, 0},
		{0, 4, 0, 0},
		{0, 4, 0, 0},
		{4, 4, 0, 0},
	}

	ExtendedPieces[3].Name = "PentoN"
	ExtendedPieces[3].Shape = [][]int{
		{6, 6, 0, 0},
		{0, 6, 6, 6},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[4].Name = "PentoNb"
	ExtendedPieces[4].Shape = [][]int{
		{5, 5, 5, 0},
		{0, 0, 5, 5},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[5].Name = "PentoP"
	ExtendedPieces[5].Shape = [][]int{
		{1, 1, 0, 0},
		{1, 1, 0, 0},
		{1, 0, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[6].Name = "PentoPb"
	ExtendedPieces[6].Shape = [][]int{
		{1, 1, 0, 0},
		{1, 1, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[7].Name = "PentoT"
	ExtendedPieces[7].Shape = [][]int{
		{2, 2, 2, 0},
		{0, 2, 0, 0},
		{0, 2, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[8].Name = "PentoU"
	ExtendedPieces[8].Shape = [][]int{
		{2, 0, 2, 0},
		{2, 2, 2, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[9].Name = "PentoV"
	ExtendedPieces[9].Shape = [][]int{
		{3, 0, 0, 0},
		{3, 0, 0, 0},
		{3, 3, 3, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[10].Name = "PentoW"
	ExtendedPieces[10].Shape = [][]int{
		{5, 0, 0, 0},
		{5, 5, 0, 0},
		{0, 5, 5, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[11].Name = "PentoX"
	ExtendedPieces[11].Shape = [][]int{
		{0, 4, 0, 0},
		{4, 4, 4, 0},
		{0, 4, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[12].Name = "PentoY"
	ExtendedPieces[12].Shape = [][]int{
		{0, 2, 0, 0},
		{2, 2, -1, 0},
		{0, 2, 0, 0},
		{0, 2, 0, 0},
	}

	ExtendedPieces[13].Name = "PentoYb"
	ExtendedPieces[13].Shape = [][]int{
		{0, 2, 0, 0},
		{-1, 2, 2, 0},
		{0, 2, 0, 0},
		{0, 2, 0, 0},
	}

	ExtendedPieces[14].Name = "PentoZ"
	ExtendedPieces[14].Shape = [][]int{
		{6, 6, 0, 0},
		{0, 6, 0, 0},
		{0, 6, 6, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[15].Name = "TinyO"
	ExtendedPieces[15].Shape = [][]int{
		{1, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[16].Name = "TinyI"
	ExtendedPieces[16].Shape = [][]int{
		{7, 0, 0, 0},
		{7, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[17].Name = "SmallL"
	ExtendedPieces[17].Shape = [][]int{
		{3, 0, 0, 0},
		{3, 3, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	ExtendedPieces[18].Name = "SmallI"
	ExtendedPieces[18].Shape = [][]int{
		{7, 0, 0, 0},
		{7, 0, 0, 0},
		{7, 0, 0, 0},
		{0, 0, 0, 0},
	}

	return ExtendedPieces
}
