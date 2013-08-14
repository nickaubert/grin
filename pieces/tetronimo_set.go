package tetronimo_set

var BasicO = [][]int{
	{ 1 , 1 , 0 , 0 } ,
	{ 1 , 1 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
}

var BasicT = [][]int{
	{ 2 , 2 , 2 , 0 } ,
	{ 0 , 2 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
}

var BasicL = [][]int{
	{ 3 , 0 , 0 , 0 } ,
	{ 3 , 0 , 0 , 0 } ,
	{ 3 , 3 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
}

var BasicJ = [][]int{
	{ 0 , 4 , 0 , 0 } ,
	{ 0 , 4 , 0 , 0 } ,
	{ 4 , 4 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
}

var BasicS = [][]int{
	{ 0 , 5 , 5 , 0 } ,
	{ 5 , 5 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
}

var BasicZ = [][]int{
	{ 6 , 6 , 0 , 0 } ,
	{ 0 , 6 , 6 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
}

var BasicI = [][]int{ 
	// includes negative phantom blocks to help rotation
	{ 0 , 7 , 0 , 0 } ,
	{ -1 , 7 , 0 , 0 } ,
	{ 0 , 7 , -1 , 0 } ,
	{ 0 , 7 , 0 , 0 } ,
}

var HugeL = [][]int{
	{ 3 , 0 , 0 , 0 } ,
	{ 3 , 0 , 0 , 0 } ,
	{ 3 , 0 , 0 , 0 } ,
	{ 3 , 3 , 3 , 0 } ,
}

var HugeJ = [][]int{
	{ 0 , 0 , 0 , 4 } ,
	{ 0 , 0 , 0 , 4 } ,
	{ 0 , 0 , 0 , 4 } ,
	{ 0 , 4 , 4 , 4 } ,
}

var HugeU = [][]int{
	{ 2 , 0 , 0 , 2 } ,
	{ 2 , 0 , 0 , 2 } ,
	{ 2 , 0 , 0 , 2 } ,
	{ 2 , 2 , 2 , 2 } ,
}

var PentoU = [][]int{
	{ 2 , 0 , 2 , 0 } ,
	{ 2 , 2 , 2 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
	{ 0 , 0 , 0 , 0 } ,
}
