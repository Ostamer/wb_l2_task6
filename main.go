// man grep - описание и основные параметры:
//
// grep ищет строки, соответствующие шаблону, и печатает их.
// Основные параметры:
// -A n: печатает n строк после совпадения
// -B n: печатает n строк до совпадения
// -C n: печатает n строк вокруг совпадения
// -c: печатает количество совпавших строк
// -i: игнорирует регистр при поиске
// -v: инвертирует совпадение, печатает строки, которые не соответствуют шаблону
// -F: выполняет точное совпадение со строкой
// -n: печатает номера строк с совпадениями

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Определение структуры для хранения параметров
type Params struct {
	After          int
	Before         int
	AfterAndBefore int
	Count          bool
	IgnoreCase     bool
	Invert         bool
	Fixed          bool
	LineNum        bool
	Pattern        string
	File           string
}

// Функция для обработки введеной строки
func setParams() Params {
	params := Params{}

	// Определение параметров
	flag.IntVar(&params.After, "A", 0, "Print N lines after match")
	flag.IntVar(&params.Before, "B", 0, "Print N lines before match")
	flag.IntVar(&params.AfterAndBefore, "C", 0, "Print N lines around match (A+B)")
	flag.BoolVar(&params.Count, "c", false, "Count matching lines")
	flag.BoolVar(&params.IgnoreCase, "i", false, "Ignore case")
	flag.BoolVar(&params.Invert, "v", false, "Invert match")
	flag.BoolVar(&params.Fixed, "F", false, "Fixed string match")
	flag.BoolVar(&params.LineNum, "n", false, "Print line numbers")

	flag.Parse()

	// Проверка является ли введенное значение в необходимом формате
	if flag.NArg() < 1 {
		fmt.Println("Введите параметры согласно примеру: название_файла параметры строка_для_обрботки файл_для_обработки")
		os.Exit(1)
	}

	// Сохранения в параметр введеную строку для паттерна
	params.Pattern = flag.Arg(0)

	// Сохранение в параметр путь до указанного файла
	if flag.NArg() > 1 {
		params.File = flag.Arg(1)
	} else {
		params.File = ""
	}

	// Установка в параметры для вывода до и после введенного паттерна, при вводе с обоих сторон
	if params.AfterAndBefore > 0 {
		params.After = params.AfterAndBefore
		params.Before = params.AfterAndBefore
	}

	return params
}

// Фукнция для чтения файла
func readFile(fileName string) ([]string, error) {
	var inputFile *os.File
	var err error

	// Чтение указанного файла иначе чтение указанной строки
	if fileName != "" {
		inputFile, err = os.Open(fileName)
		if err != nil {
			return nil, err
		}
		defer inputFile.Close()
	} else {
		inputFile = os.Stdin
	}

	// Сохранение строк файла в переменную
	scanner := bufio.NewScanner(inputFile)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// Функция для составления фукнции для обработки файла с учетом введеных параметров
func compileFile(params Params) func(string) bool {
	// Если указан параметр Fixed, создается функция, которая ищет вхождение фиксированной введеного паттерна строки.
	// Иначе ищем при помощи регулярного выражения
	if params.Fixed {
		return func(line string) bool {
			// Если указан параметр IgnoreCase, строка приводится к нижнему регистру.
			if params.IgnoreCase {
				line = strings.ToLower(line)
				params.Pattern = strings.ToLower(params.Pattern)
			}
			// Проверка, содержит ли строка указанный шаблон.
			return strings.Contains(line, params.Pattern)
		}
	} else {
		re := regexp.MustCompile(params.Pattern)
		return func(line string) bool {
			// Если указан параметр IgnoreCase, строка приводится к нижнему регистру.
			if params.IgnoreCase {
				line = strings.ToLower(line)
			}
			// Проверка, соответствует ли строка регулярному выражению.
			return re.MatchString(line)
		}
	}
}

// Фукнция для прохода по файлу с функцией
func processLines(lines []string, compile func(string) bool, params Params) {
	// Список совпадений хранит индексы строк, подходящих под условия.
	result := []int{}

	// Проход по каждой строке файла.
	for i, line := range lines {
		// Если результат функции matcher(line) не совпадает с флагом Invert, добавляем индекс строки в matches.
		if compile(line) != params.Invert {
			result = append(result, i)
		}
	}

	// Если указан параметр Count, выводим количество совпадений и завершаем выполнение.
	if params.Count {
		fmt.Println(len(result))
		return
	}

	// Словарь для вывода только уникальных строк
	existing := map[int]bool{}

	// Проход по индексам строк, которые совпали с условиями.
	for _, line := range result {
		// Определяем диапазон строк для вывода (до и после совпадения).
		start := max(0, line-params.Before)
		end := min(len(lines), line+params.After+1)
		for i := start; i < end; i++ {
			// Если строка еще не была напечатана, выводим ее.
			if !existing[i] {
				if params.LineNum {
					// Если включен LineNum, выводим номер строки перед содержимым.
					fmt.Printf("%d:", i+1)
				}
				// Вывод строки
				fmt.Println(lines[i])
				// Отмечаем строку как напечатанную.
				existing[i] = true
			}
		}
	}
}

// Функция для поиска макс значения между двумя элементами
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Фукнция для поиска мин значения между двумя элементами
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Основная функция
func main() {
	// Обрабатываем введенную строку и разбиваем ее на параметры
	params := setParams()

	// Чтение файла
	lines, err := readFile(params.File)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Компилирование функции с указанными параметрами
	compile := compileFile(params)

	// Обработка файла с созданной фукнции
	processLines(lines, compile, params)
}
