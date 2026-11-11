package tray

// garagePixel is the Windows icon's shape on a 32×32 grid: a roof, walls and a slatted door, so
// the notification area shows a garage in the state's colour (macOS uses SF Symbols instead).
func garagePixel(x, y int) bool {
	fx, fy := float64(x)+0.5, float64(y)+0.5
	if fy >= 3 && fy < 14 { // roof: apex at (16, 3), eaves from 1 to 31
		return fx-16 <= (fy-3)*15/11 && 16-fx <= (fy-3)*15/11
	}
	if fy < 14 || fy > 29 || fx < 4 || fx > 28 {
		return false
	}
	if fx > 9 && fx < 23 && fy > 16 { // the door: 2-pixel slats with 2-pixel gaps
		r := (y - 17) % 4
		return r == 0 || r == 1
	}
	return true
}
