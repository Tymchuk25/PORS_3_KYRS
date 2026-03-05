package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// reverse — розвертає слайс задом наперед
// [1, 4, 5, 3] → [3, 5, 4, 1]
func reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < len(s)/2; i++ {
		s[i], s[j-i] = s[j-i], s[i] //s[0], s[3] = s[3], s[0]
	}
	return s
}

// allDifferentNeighbors — перевіряє чи всі сусідні елементи різні
// [3, 5, 4, 1] → true (всі різні), [3, 3, 4, 1] → false (3==3)
func allDifferentNeighbors(s []int) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == s[i+1] {
			return false
		}
	}
	return true
}

// digitSum — повертає суму всіх елементів слайсу
// [3, 5, 4, 1] → 13
func digitSum(s []int) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += s[i]
	}
	return sum
}

// isNotEqualToABC — перевіряє що sum не дорівнює жодному з a, b, c
func isNotEqualToABC(sum int, a int, b int, c int) bool {
	return sum != a && sum != b && sum != c
}

// countInRange — рахує кількість "хороших" чисел у діапазоні [from, to]
// Це той самий код що був в main, просто обгорнутий у функцію
func countInRange(from, to, n, a, b, c int) int {
	count := 0
	for num := from; num <= to; num++ {
		ds := []int{}
		digit := num
		for i := 0; i < n; i++ {
			ds = append(ds, digit%10)
			digit /= 10
		}
		ds = reverse(ds)

		if allDifferentNeighbors(ds) && isNotEqualToABC(digitSum(ds), a, b, c) {
			count++
		}
	}
	return count
}

// workerResult — результат роботи однієї горутіни
type workerResult struct {
	id    int
	count int
	time  time.Duration
}

// parResult — результат паралельного запуску з N горутінами
type parResult struct {
	numWorkers int
	time       time.Duration
}

// countInRangeGoroutine — те саме, але результат відправляє в канал
// Це обгортка для countInRange, щоб працювало з go
func countInRangeGoroutine(id, from, to, n, a, b, c int, ch chan workerResult) {
	start := time.Now()
	count := countInRange(from, to, n, a, b, c)
	elapsed := time.Since(start)
	ch <- workerResult{id: id, count: count, time: elapsed}
}

// generateInput — генерує випадкові вхідні дані і зберігає у файл
func generateInput(filename string) {
	n := rand.Intn(7) + 2 // n від 2 до 8
	maxSum := n * 9       // максимально можлива сума цифр
	a := rand.Intn(maxSum) + 1
	b := rand.Intn(maxSum) + 1
	c := rand.Intn(maxSum) + 1

	file, _ := os.Create(filename)
	fmt.Fprintf(file, "%d %d %d %d\n", n, a, b, c)
	file.Close()

	fmt.Printf("Згенеровано вхідні дані: n=%d, a=%d, b=%d, c=%d → %s\n", n, a, b, c, filename)
}

// readInput — читає n, a, b, c з файлу
func readInput(filename string) (int, int, int, int) {
	file, _ := os.Open(filename)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	parts := strings.Fields(scanner.Text())

	n, err := strconv.Atoi(parts[0]) //трансформуємо рядок у число, наприклад "3" → 3
	if err != nil {
		fmt.Println("Помилка:", err)
		return 0, 0, 0, 0
	}
	a, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("Помилка:", err)
		return 0, 0, 0, 0
	}
	b, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Println("Помилка:", err)
		return 0, 0, 0, 0
	}
	c, err := strconv.Atoi(parts[3])
	if err != nil {
		fmt.Println("Помилка:", err)
		return 0, 0, 0, 0
	}
	return n, a, b, c
}

// writeOutput — записує результати у файл
func writeOutput(filename string, n, a, b, c, result int, seqTime time.Duration, parResults []parResult) {
	file, _ := os.Create(filename)
	defer file.Close()

	fmt.Fprintf(file, "Вхідні дані: n=%d, a=%d, b=%d, c=%d\n", n, a, b, c)
	fmt.Fprintf(file, "Результат: %d\n", result)
	fmt.Fprintf(file, "Послідовно: %v\n", seqTime)
	for _, pr := range parResults {
		fmt.Fprintf(file, "Паралельно (%d горутін): %v\n", pr.numWorkers, pr.time)
	}
}

func main() {
	// Якщо передано аргумент --generate — генеруємо вхідні дані
	// Використання: ./lab1 --generate
	if len(os.Args) > 1 && os.Args[1] == "--generate" {
		generateInput("input.txt")
		return
	}

	var n, a, b, c int

	if len(os.Args) == 5 {
		// Використання: ./lab1 n a b c
		n, _ = strconv.Atoi(os.Args[1])
		a, _ = strconv.Atoi(os.Args[2])
		b, _ = strconv.Atoi(os.Args[3])
		c, _ = strconv.Atoi(os.Args[4])
	} else {
		// Читаємо вхідні дані з файлу
		// Формат input.txt: один рядок "n a b c"
		n, a, b, c = readInput("input.txt")
	}

	// Рахуємо діапазон n-цифрових чисел
	// n=2: from=10, to=99
	// n=3: from=100, to=999
	from := 1
	for i := 0; i < n-1; i++ {
		from *= 10
	}
	to := from*10 - 1

	fmt.Printf("n=%d, a=%d, b=%d, c=%d\n", n, a, b, c)
	fmt.Printf("Перебираємо числа від %d до %d\n\n", from, to)

	// === Послідовна версія ===
	start := time.Now()
	result := countInRange(from, to, n, a, b, c)
	seqTime := time.Since(start)

	fmt.Printf("Послідовно: %d (час: %v)\n\n", result, seqTime)

	// === Паралельна версія для різної кількості горутін ===
	workerCounts := []int{2, 4, 8}
	var parResults []parResult

	for _, numWorkers := range workerCounts {
		ch := make(chan workerResult, numWorkers) // канал для отримання результатів від горутін

		totalNumbers := to - from + 1          // загальна кількість чисел у діапазоні
		chunkSize := totalNumbers / numWorkers // розмір частини для кожної горутіни

		start = time.Now() // початок відліку
		for w := 0; w < numWorkers; w++ {
			wFrom := from + w*chunkSize  // початок діапазону для цього воркера
			wTo := wFrom + chunkSize - 1 // кінець діапазону для цього воркера
			if w == numWorkers-1 {
				wTo = to // останній воркер забирає залишок
			}
			go countInRangeGoroutine(w, wFrom, wTo, n, a, b, c, ch) // запускаємо горутіну
		}

		// Збираємо результати з каналу — кожна горутіна відправить свій результат
		parallelResult := 0
		for w := 0; w < numWorkers; w++ {
			res := <-ch
			//			fmt.Printf("  Горутіна %d: %d (час: %v)\n", res.id, res.count, res.time)
			parallelResult += res.count
		}
		parTime := time.Since(start) // кінець відліку

		fmt.Printf("Паралельно (%d горутін): %d (час: %v)\n\n", numWorkers, parallelResult, parTime)
		parResults = append(parResults, parResult{numWorkers: numWorkers, time: parTime})
	}

	// Записуємо результати у файл
	writeOutput("output.txt", n, a, b, c, result, seqTime, parResults)
	fmt.Println("Результати збережено у output.txt")
}
