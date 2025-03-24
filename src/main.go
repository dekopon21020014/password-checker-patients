package main

import (
	"fmt"
	"log"
	"os"
	"password-checker-tui/table"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/rivo/tview"
)

var (
	wg            sync.WaitGroup
	app           *tview.Application
	layout        *tview.Flex
	password      string
	currentDir    string
	patientsTable *table.PatientsTable

	// 各種UI
	passwordForm *tview.Form
	sqliteList   *tview.List
	logView      *tview.TextView
)

// ** メイン関数 **
func main() {
	// .envファイルを読み込む
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	setupLogger()

	for { // ユーザが"終了"を選択するまでループ
		initializeTUI()
		wg = sync.WaitGroup{}
		if err := app.Run(); err != nil {
			log.Fatalf("failed to start app: %v", err)
		}
		wg.Wait()
	}
}

// ** ログ設定 **
func setupLogger() {
	// 保存フォルダが存在しない場合は作成
	logFileDir := os.Getenv("LOG_FILE_DIR")
	if _, err := os.Stat(logFileDir); os.IsNotExist(err) {
		err := os.MkdirAll(logFileDir, 0755) // フォルダ作成
		if err != nil {
			log.Fatalf("保存フォルダの作成に失敗: %v", err)
		}
	}

	logFile, err := os.OpenFile(
		filepath.Join(logFileDir, os.Getenv("LOG_FILE_NAME")),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatalf("ログファイル作成失敗: %v", err)
	}
	log.SetOutput(logFile)
}

// ** TUIの初期化 **
func initializeTUI() {
	app = tview.NewApplication()
	// currentDir, _ = os.Getwd()
	currentDir = os.Getenv("CURRENT_DIR")

	// 各画面を作成
	passwordForm = createPasswordForm()
	sqliteList = createSqlitelList()
	logView = createLogView()

	// **最初はパスワード画面を表示**
	layout = tview.NewFlex().
		AddItem(passwordForm, 0, 1, true).
		AddItem(sqliteList, 0, 1, false)

	app.SetRoot(layout, true)
}

// ** パスワード入力フォーム **
func createPasswordForm() *tview.Form {
	form := tview.NewForm().
		AddPasswordField("パスワード:", "", 20, '*', func(text string) {
			password = text
		}).
		AddButton("次へ", func() {
			updateSqliteList()
		}).
		AddButton("終了", func() {
			app.Stop()
			os.Exit(0)
		})

	form.SetBorder(true).SetTitle("1. パスワード入力")
	return form
}

func createSqlitelList() *tview.List {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle("2. データベースファイル(.sqlite)を選択")
	return list
}

// ** ログ画面 **
func createLogView() *tview.TextView {
	logView := tview.NewTextView().SetDynamicColors(true)
	logView.SetBorder(true).SetTitle("ログ")
	return logView
}

// ** sqlite選択リストを更新 **
func updateSqliteList() {
	go func() { // 非同期で処理
		sqliteList.Clear()

		entries, err := os.ReadDir(currentDir)
		if err != nil {
			logView.SetText("ディレクトリ読み取り失敗: " + err.Error())
			return
		}
		for _, entry := range entries {
			filePath := filepath.Join(currentDir, entry.Name())

			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				sqliteList.AddItem("[DIR] "+entry.Name(), "", 0, func() {
					currentDir = filePath
					updateSqliteList()
				})
			} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".sqlite") {
				sqliteList.AddItem(entry.Name(), "", 0, func() {
					patientsTable = table.GetDatabase(filepath.Join(currentDir, entry.Name()))
					defer patientsTable.DB.Close()
					result, err := patientsTable.CheckPassword(password)
					var msg string
					if err != nil {
						log.Println("Error in CheckPassword: ", err)
						msg = fmt.Sprintf("Error in CheckPassword: %v", err)
					} else {
						msg = createResultMessage(result)
					}
					showCompletionMenu(msg)
				})
			}
		}

		// 親ディレクトリ (..) を追加
		parentDir := filepath.Dir(currentDir)
		sqliteList.AddItem("[DIR] 前のフォルダに戻る", "", 0, func() {
			currentDir = parentDir
			updateSqliteList()
		})

		// **[追加] 画面を更新 & フォーカス移動**
		app.SetFocus(sqliteList)
		app.Draw()
	}()
}

func createResultMessage(result map[string]map[string]int) string {
	msg := "パスワード照合結果:\n"
	if len(result) == 0 {
		return "チェック結果: データがありません。"
	}

	var dates []string
	for date := range result {
		dates = append(dates, date)
	}

	// 日付順にソート
	sort.Slice(dates, func(i, j int) bool {
		t1, _ := time.Parse("2006/01/02", dates[i])
		t2, _ := time.Parse("2006/01/02", dates[j])
		return t1.Before(t2) // 昇順ソート
	})

	// ソートされた順番でメッセージを作成
	for _, date := range dates {
		counts := result[date]
		msg += fmt.Sprintf("%s - 一致: %d, 不一致: %d\n", date, counts["一致"], counts["不一致"])
	}

	return msg
}

// ** 処理完了メニュー **
func showCompletionMenu(msg string) {
	app.Stop()
	app = tview.NewApplication()
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"続ける", "終了"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "終了" {
				app.Stop()
				os.Exit(0)
			} else {
				app.Stop()
			}
		})

	if err := app.SetRoot(modal, true).Run(); err != nil {
		log.Fatalf("アプリケーションエラー: %v", err)
	}
}
