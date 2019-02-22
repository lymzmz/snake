package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	RIGHT = 100
	LEFT  = 97
	UP    = 119
	DOWN  = 115
)

type coord struct {
	X       int
	Y       int
	S       string
	HasFood chan bool
}

func main() {
	fmt.Println("Please choose which one:")
	fmt.Println("1.rain")
	fmt.Println("2.snake")

	choose := 0
	fmt.Scan(&choose)

	if choose != 1 && choose != 2 {
		fmt.Println("I don't know you want to do!")
		return
	}

	switch choose {
	case 1:
		NewRain().Start()
	case 2:
		NewSnake().Start()
	}
}

type rain struct {
	Body   map[string]coord
	Food   coord
	Width  int
	Height int
	Speed  time.Duration
	Score  int64
	sync.RWMutex
}

func NewRain() *rain {
	return &rain{
		Body:   make(map[string]coord),
		Width:  20,
		Height: 20,
		Speed:  100,
	}
}

func (r *rain) Start() {
	var speed time.Duration
	fmt.Print("please set speed(ms):")
	fmt.Scan(&speed)
	r.Speed = speed

	defer func() {
		if err := recover(); err != nil {
			fmt.Print(err)
		}
	}()
	go r.Keyboard()
	go r.Gfood()
	go r.Move()
	for {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}

		fmt.Println("score:", r.Score)
		fmt.Print(r.Draw())

		time.Sleep(time.Millisecond * r.Speed)
	}
}

func (r *rain) Keyboard() {
	for {
		c := getch()
		for index, item := range r.Body {
			if item.S == string(c) {
				r.Score++
				r.Lock()
				delete(r.Body, index)
				r.Unlock()
			}
		}
	}
}

func (r *rain) Move() {
	for {
		r.Lock()
		for index, item := range r.Body {
			item.Y++
			if item.Y >= r.Height {
				delete(r.Body, index)
			} else {
				r.Body[index] = item
			}
		}
		r.Unlock()
		time.Sleep(time.Millisecond * r.Speed)
	}
}

func (r *rain) Gfood() {
	for {
		str := []byte("abcdefghjiklmnopqrstuvwxyz0123456789")
		r.Food.X = rand.Intn(r.Width)
		r.Food.Y = 0
		r.Food.S = string(str[rand.Intn(len(str))])
		r.Lock()
		r.Body[r.Food.S] = r.Food
		r.Unlock()
		time.Sleep(time.Millisecond * r.Speed)
	}
}

func (r *rain) Draw() string {
	r.Lock()
	defer r.Unlock()

	str := ""
	str += string('\r') + strings.Repeat("-", r.Width+2) + string('\n')
	for i := 0; i < r.Height; i++ {
		str += string('\r') + "-"
		for j := 0; j < r.Width; j++ {
			keyword := ""
			for k := range r.Body {
				if r.Body[k].X == j && r.Body[k].Y == i {
					keyword = r.Body[k].S
					break
				}
			}
			if keyword != "" {
				str += keyword
			} else {
				str += " "
			}
		}
		str += "-" + string('\n')
	}
	str += string('\r') + strings.Repeat("-", r.Width+2) + string('\n')

	return str
}

type snake struct {
	Body   []coord
	Food   coord
	Width  int
	Height int
	Speed  time.Duration
	Direct byte
	Pause  bool
	sync.WaitGroup
}

func NewSnake() *snake {
	food := coord{
		HasFood: make(chan bool, 1),
	}
	food.HasFood <- true
	return &snake{
		Body:   []coord{coord{X: 0, Y: 0}},
		Food:   food,
		Speed:  350,
		Width:  20,
		Height: 7,
		Direct: RIGHT,
	}
}

func (s *snake) Start() {
	var speed time.Duration
	fmt.Print("please set speed(ms):")
	fmt.Scan(&speed)
	s.Speed = speed

	go s.Gfood()
	go s.Keyboard()
	defer func() {
		if err := recover(); err != nil {
			fmt.Print(err)
			s.Fail()
		}
	}()
	for {
		s.Wait()
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
		s = s.Move()
		fmt.Println(len(s.Body), s.Body[0].X, s.Body[0].Y)
		fmt.Print(s.Draw())

		time.Sleep(time.Millisecond * s.Speed)
	}
}

func (s *snake) Keyboard() {
	for {
		c := getch()
		switch c[0] {
		case 32:
			if s.Pause == true {
				s.Done()
				s.Pause = false
			} else {
				fmt.Print("pause.")
				s.Add(1)
				s.Pause = true
			}
		default:
			s.Direct = c[0]
		}
	}
}

func (s *snake) Gfood() {
HERE:
	for _ = range s.Food.HasFood {
		s.Food.X = rand.Intn(s.Width - 1)
		s.Food.Y = rand.Intn(s.Height - 1)

		for i, j := 0, len(s.Body); i < j; i++ {
			if s.Body[i].X == s.Food.X && s.Body[i].Y == s.Food.Y {
				s.Food.HasFood <- true
				goto HERE
			}
		}
		return
	}
}

func (s *snake) move(x, y int) []coord {
	tmpSnake := make([]coord, len(s.Body))
	tmpSnake[0].X, tmpSnake[0].Y = x, y

	for i, j := 0, len(s.Body)-1; i < j; i++ {
		tmpSnake[i+1] = s.Body[i]
	}
	return tmpSnake
}

func (s *snake) Move() *snake {
	switch s.Direct {
	case RIGHT:
		s.Body = s.move(s.Body[0].X+1, s.Body[0].Y)
	case LEFT:
		s.Body = s.move(s.Body[0].X-1, s.Body[0].Y)
	case UP:
		s.Body = s.move(s.Body[0].X, s.Body[0].Y-1)
	case DOWN:
		s.Body = s.move(s.Body[0].X, s.Body[0].Y+1)
	}
	if s.Body[0].X < 0 || s.Body[0].X > s.Width-1 || s.Body[0].Y < 0 || s.Body[0].Y > s.Height-1 {
		s.Fail()
	}
	for k, l := 1, len(s.Body); k < l; k++ {
		if s.Body[k].X == s.Body[0].X && s.Body[k].Y == s.Body[0].Y {
			s.Fail()
			break
		}
	}
	if s.Food.X == s.Body[0].X && s.Food.Y == s.Body[0].Y {
		s.Food.HasFood <- false
		s.Gfood()
		s.Body = append(s.Body, s.graw())
		//s.Move()
	}

	return s
}

func (s *snake) graw() coord {
	tmpBody := coord{}
	l := len(s.Body)
	if l == 1 {
		last := s.Body[0]
		switch s.Direct {
		case UP:
			tmpBody.X = last.X
			tmpBody.Y = last.Y + 1
		case DOWN:
			tmpBody.X = last.X
			tmpBody.Y = last.Y - 1
		case LEFT:
			tmpBody.X = last.X + 1
			tmpBody.Y = last.Y
		case RIGHT:
			tmpBody.X = last.X - 1
			tmpBody.Y = last.Y
		}
	} else {
		last := s.Body[l-1]
		last2 := s.Body[l-2]

		if last.X > last2.X {
			tmpBody.X = last.X + 1
			tmpBody.Y = last.Y
		} else if last.X < last2.X {
			tmpBody.X = last.X - 1
			tmpBody.Y = last.Y
		} else if last.Y > last2.Y {
			tmpBody.X = last.X
			tmpBody.Y = last.Y + 1
		} else if last.Y < last2.Y {
			tmpBody.X = last.X
			tmpBody.Y = last.Y - 1
		}
	}
	return tmpBody
}

func (s *snake) Draw() string {
	str := ""
	str += string('\r') + strings.Repeat("-", s.Width+2) + string('\n')
	for i := 0; i < s.Height; i++ {
		str += string('\r') + "-"
		for j := 0; j < s.Width; j++ {
			snakeBody := false
			for k, l := 0, len(s.Body); k < l; k++ {
				if s.Body[k].X == j && s.Body[k].Y == i {
					snakeBody = true
					break
				}
			}
			if snakeBody {
				str += "*"
			} else if s.Food.Y == i && s.Food.X == j {
				str += "*"
			} else {
				str += " "
			}
		}
		str += "-" + string('\n')
	}
	str += string('\r') + strings.Repeat("-", s.Width+2) + string('\n')

	return str
}

func (s *snake) Fail() {
	fmt.Print("fail!")
	s.Add(1)
	s.Reset()
	s.Pause = true
}

func (s *snake) Reset() {
	s.Body = []coord{coord{X: 0, Y: 0}}
	s.Direct = RIGHT
}