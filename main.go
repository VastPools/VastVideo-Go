package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"runtime"
	"sync"
	"vastproxy-go/components"
	"vastproxy-go/utils"
)

//go:embed html/index_mobile.html html/about.html html/type_mapping.html html/sources_manage.html test_sources.html
var htmlContent embed.FS

//go:embed config/config.ini config/sources.json
var ConfigContent embed.FS

//go:embed html/check_sources.html
var checkSourcesHTML embed.FS

// ScorpioSource 结构体
// 用于解析和保存 scorpio.json 的每个资源
// 只在 main.go 内部使用

type ScorpioSource struct {
	Name          string `json:"name"`
	API           string `json:"api"`
	LastCheckTime int64  `json:"last_check_time,omitempty"`
	IsValid       *bool  `json:"is_valid,omitempty"`
}

const ScorpioJsonPath = "config/scorpio.json"

// 读取scorpio.json
func LoadScorpioSources() ([]*ScorpioSource, error) {
	f, err := os.Open(ScorpioJsonPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var sources []*ScorpioSource
	dec := json.NewDecoder(f)
	if err := dec.Decode(&sources); err != nil {
		return nil, err
	}
	return sources, nil
}

// 保存scorpio.json
func SaveScorpioSources(sources []*ScorpioSource) error {
	data, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ScorpioJsonPath, data, 0644)
}

// 检查单个资源，返回是否可用、消息、JSON内容、响应时间（毫秒）
func CheckSourceAPIWithBody(api string) (bool, string, interface{}, int64) {
	start := time.Now()
	client := &http.Client{Timeout: 15 * time.Second} // 增加最大等待时间
	resp, err := client.Get(api)
	cost := time.Since(start).Milliseconds()
	if err != nil {
		return false, err.Error(), nil, cost
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		var result interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&result); err == nil {
			return true, "ok", result, cost
		}
		return true, "ok (非JSON)", nil, cost
	}
	return false, fmt.Sprintf("HTTP %d", resp.StatusCode), nil, cost
}

// 资源检测页面
func checkSourcesPageHandler(w http.ResponseWriter, r *http.Request) {
	htmlBytes, err := checkSourcesHTML.ReadFile("html/check_sources.html")
	if err != nil {
		http.Error(w, "无法读取HTML模板", 500)
		return
	}
	sources, err := LoadScorpioSources()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlBytes)
	fmt.Fprintf(w, `<script>\nwindow._scorpio_sources = %s;\n</script>`, mustJson(sources))
}

func mustJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// SSE流接口
func checkSourcesStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sources, err := LoadScorpioSources()
	if err != nil {
		fmt.Fprintf(w, "data: {\"msg\":\"读取scorpio.json失败\"}\n\n")
		w.(http.Flusher).Flush()
		return
	}

	var mu sync.Mutex
	for i, src := range sources {
		go func(idx int, s *ScorpioSource) {
			isValid, msg, result, cost := CheckSourceAPIWithBody(s.API)
			t := time.Now().Unix()
			// 响应时间评级
			level := "快"
			if cost > 8000 {
				level = "慢"
			} else if cost > 3000 {
				level = "中"
			}
			mu.Lock()
			s.LastCheckTime = t
			s.IsValid = &isValid
			SaveScorpioSources(sources)
			mu.Unlock()
			res := map[string]interface{}{
				"index":           idx,
				"name":            s.Name,
				"api":             s.API,
				"last_check_time": t,
				"is_valid":        isValid,
				"msg":             msg,
				"response_time":   cost,
				"response_level":  level,
			}
			if isValid && result != nil {
				res["result_json"] = result
			}
			b, _ := json.Marshal(res)
			fmt.Fprintf(w, "data: %s\n\n", b)
			w.(http.Flusher).Flush()
		}(i, src)
	}

	// 等待所有goroutine完成
	time.Sleep(time.Duration(len(sources)) * 16 * time.Second)
}

// 检查单个资源API（前端逐个调用）
func HandleCheckSourceAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api := r.URL.Query().Get("api")
	if api == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Missing api parameter",
		})
		return
	}
	isValid, msg, result, cost := CheckSourceAPIWithBody(api)
	level := "快"
	if cost > 8000 {
		level = "慢"
	} else if cost > 3000 {
		level = "中"
	}
	resp := map[string]interface{}{
		"success":        true,
		"is_valid":       isValid && result != nil,
		"msg":            msg,
		"response_time":  cost,
		"response_level": level,
	}
	if isValid && result != nil {
		resp["result_json"] = result
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

var GlobalConfig *utils.Config

func main() {
	// 加载配置文件
	if err := LoadConfig(); err != nil {
		log.Fatalf("❌ 加载配置文件失败: %v", err)
	}

	// 初始化视频源配置
	sourcesConfig := components.NewSourcesConfig()
	sourcesData, err := ConfigContent.ReadFile("config/sources.json")
	if err != nil {
		log.Fatalf("❌ 读取视频源配置文件失败: %v", err)
	}
	if err := sourcesConfig.LoadFromConfigFile(sourcesData); err != nil {
		log.Fatalf("❌ 加载视频源配置失败: %v", err)
	}
	log.Printf("✅ 视频源配置加载成功，共 %d 个源", len(sourcesConfig.GetSources()))

	// 初始化类型映射管理器
	typeMappingManager := components.NewTypeMappingManager("config/type_mapping.json")
	if err := typeMappingManager.LoadConfig(); err != nil {
		log.Printf("⚠️ 加载类型映射配置失败: %v，将使用默认配置", err)
		// 如果配置文件不存在，可以创建一个默认配置
	} else {
		log.Printf("✅ 类型映射配置加载成功")
	}

	// 将类型映射管理器设置到视频源配置中
	sourcesConfig.SetTypeMappingManager(typeMappingManager)

	// 定义命令行参数
	var (
		port = flag.String("port", GlobalConfig.Server.Port, "服务端口")
	)
	flag.Parse()

	// 设置日志输出
	var outputs []io.Writer
	if GlobalConfig.Logging.ConsoleOutput {
		outputs = append(outputs, os.Stdout)
	}
	if GlobalConfig.Logging.FileOutput {
		logFile, err := os.OpenFile(GlobalConfig.Logging.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("无法打开日志文件: %v", err)
		}
		outputs = append(outputs, logFile)
	}

	if len(outputs) > 0 {
		log.SetOutput(io.MultiWriter(outputs...))
	}

	// 检查并处理端口占用
	log.Printf("🔍 检查端口 %s 是否可用...", *port)
	if err := checkAndKillPortProcess(*port); err != nil {
		log.Fatalf("❌ 端口检查失败: %v", err)
	}

	// 注册路由
	if GlobalConfig.Features.ProxyService {
		http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
			components.ProxyHandler(w, r, GlobalConfig)
		})
	}
	if GlobalConfig.Features.HealthCheck {
		http.HandleFunc("/health", healthHandler)
	}
	if GlobalConfig.Features.InfoPage {
		http.HandleFunc("/info", infoHandler)
		http.HandleFunc("/mobile", mobileHandler)
		http.HandleFunc("/about", aboutHandler)
		http.HandleFunc("/about.html", aboutHandler)
		http.HandleFunc("/type_mapping", typeMappingHandler)
		http.HandleFunc("/sources_manage", sourcesManageHandler)
		http.HandleFunc("/test_sources", testSourcesHandler)
		http.HandleFunc("/test_sources_manage", testSourcesManageHandler)
		http.HandleFunc("/demo", demoHandler)
		http.HandleFunc("/sources_manage_preview", sourcesManagePreviewHandler)
		http.HandleFunc("/manage", manageHandler)
		http.HandleFunc("/", indexHandler)
	}
	if GlobalConfig.Features.DoubanAPI {
		http.HandleFunc("/douban", func(w http.ResponseWriter, r *http.Request) {
			components.DoubanHandler(w, r, GlobalConfig)
		})
	}

	// 添加视频源API路由
	http.HandleFunc("/api/sources", sourcesConfig.HandleSourcesAPI)
	http.HandleFunc("/api/source_search", sourcesConfig.HandleSourceSearchAPI)

	// 添加视频源管理API路由
	sourcesManageAPIHandler := components.NewSourcesManageHandler(sourcesConfig, "config/sources.json")
	http.HandleFunc("/api/sources_manage/", sourcesManageAPIHandler.HandleSourcesManageAPI)
	http.HandleFunc("/api/sources_manage", sourcesManageAPIHandler.HandleSourcesManageAPI)

	// 添加类型映射API路由
	http.HandleFunc("/api/type_mapping", typeMappingManager.HandleTypeMappingAPI)
	http.HandleFunc("/api/type_mapping/", typeMappingManager.HandleTypeMappingAPI)

	// 添加类型映射管理API路由
	http.HandleFunc("/api/type_mapping_manage", typeMappingManager.HandleTypeMappingManageAPI)
	http.HandleFunc("/api/type_mapping_manage/", typeMappingManager.HandleTypeMappingManageAPI)

	// 添加过滤配置API路由
	http.HandleFunc("/api/filter_config", filterConfigHandler)

	// 新增：资源检测页面和SSE流
	http.HandleFunc("/check_sources", checkSourcesPageHandler)
	// 移除 http.HandleFunc("/check_sources/stream", ...) 及 checkSourcesStreamHandler 相关实现

	// 添加 scorpio 源 API 路由
	http.HandleFunc("/api/scorpio_sources", components.HandleScorpioSourcesAPI)
	http.HandleFunc("/api/scorpio_sources/", components.HandleScorpioSourcesAPI)
	http.HandleFunc("/api/check_source", HandleCheckSourceAPI)

	// 获取本地IP地址
	localIP := components.GetLocalIP()

	log.Println("🚀 VastProxy-Go 代理服务启动中...")
	log.Printf("📍 服务地址: http://%s:%s", localIP, *port)
	if GlobalConfig.Features.HealthCheck {
		log.Printf("🔗 健康检查: http://%s:%s/health", GlobalConfig.Server.Host, *port)
	}
	if GlobalConfig.Features.InfoPage {
		log.Printf("📄 信息页面: http://%s:%s/info", GlobalConfig.Server.Host, *port)
		log.Printf("📱 移动端页面: http://%s:%s/mobile", GlobalConfig.Server.Host, *port)
		log.Printf("🏠 首页(移动端): http://%s:%s/", GlobalConfig.Server.Host, *port)
		log.Printf("🎯 类型映射管理: http://%s:%s/type_mapping", GlobalConfig.Server.Host, *port)
		log.Printf("🔧 视频源管理: http://%s:%s/sources_manage", GlobalConfig.Server.Host, *port)
		log.Printf("🧪 视频源管理测试: http://%s:%s/test_sources_manage", GlobalConfig.Server.Host, *port)
		log.Printf("🎬 功能演示页面: http://%s:%s/demo", GlobalConfig.Server.Host, *port)
		log.Printf("🎨 美化预览页面: http://%s:%s/sources_manage_preview", GlobalConfig.Server.Host, *port)
		log.Printf("🎛️ 统一管理控制台: http://%s:%s/manage", GlobalConfig.Server.Host, *port)
	}
	if GlobalConfig.Features.DoubanAPI {
		log.Printf("🎬 豆瓣API: http://%s:%s/douban", GlobalConfig.Server.Host, *port)
	}
	log.Printf("🎯 视频源API: http://%s:%s/api/sources", GlobalConfig.Server.Host, *port)
	log.Printf("📝 日志文件: %s", GlobalConfig.Logging.LogFile)
	log.Println(strings.Repeat("=", 50))

	// 启动服务器
	go func() {
		err := http.ListenAndServe(GlobalConfig.Server.Host+":"+*port, nil)
		if err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待一秒确保服务器启动
	time.Sleep(1 * time.Second)

	log.Printf("📱 服务器已启动，访问地址: http://%s:%s/", localIP, *port)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 保持主程序运行，等待信号或浏览器关闭
	select {
	case sig := <-sigChan:
		log.Printf("📴 收到信号 %v，正在退出...", sig)
	case <-time.After(24 * time.Hour): // 防止无限等待
		log.Println("⏰ 程序运行超时，正在退出...")
	}
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	configData, err := ConfigContent.ReadFile("config/config.ini")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	config, err := utils.LoadConfigFromData(configData)
	if err != nil {
		return err
	}

	GlobalConfig = config
	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"VastProxy-Go","version":"1.0.0","timestamp":` +
		fmt.Sprintf("%d", time.Now().Unix()) + `}`))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	// 只处理 /info 路径请求
	if r.URL.Path != "/info" {
		http.NotFound(w, r)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 读取嵌入的 HTML 文件
	content, err := htmlContent.ReadFile("html/info.html")
	if err != nil {
		log.Printf("❌ 读取嵌入的 HTML 文件失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 返回 HTML 内容
	w.Write(content)
	log.Printf("📄 返回信息页面 HTML [IP:%s]", utils.GetRequestIP(r))
}

func mobileHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/mobile" {
		http.Redirect(w, r, "/mobile", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	content, err := htmlContent.ReadFile("html/index_mobile.html")
	if err != nil {
		log.Printf("❌ 读取 html/index_mobile.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(content)
	log.Printf("📱 返回移动端页面 html/index_mobile.html [IP:%s]", utils.GetRequestIP(r))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 直接返回移动端页面，取消客户端类型判断
	content, err := htmlContent.ReadFile("html/index_mobile.html")
	if err != nil {
		log.Printf("❌ 读取 html/index_mobile.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("📱 返回移动端页面 html/index_mobile.html [IP:%s]", utils.GetRequestIP(r))
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/about" && r.URL.Path != "/about.html" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := htmlContent.ReadFile("html/about.html")
	if err != nil {
		log.Printf("❌ 读取 html/about.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("📄 返回关于页面 html/about.html [IP:%s]", utils.GetRequestIP(r))
}

func typeMappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/type_mapping" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := htmlContent.ReadFile("html/type_mapping.html")
	if err != nil {
		log.Printf("❌ 读取 html/type_mapping.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("🎯 返回类型映射管理页面 html/type_mapping.html [IP:%s]", utils.GetRequestIP(r))
}

func sourcesManageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/sources_manage" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := htmlContent.ReadFile("html/sources_manage.html")
	if err != nil {
		log.Printf("❌ 读取 html/sources_manage.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("🎯 返回视频源管理页面 html/sources_manage.html [IP:%s]", utils.GetRequestIP(r))
}

func testSourcesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/test_sources" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := htmlContent.ReadFile("test_sources.html")
	if err != nil {
		log.Printf("❌ 读取 test_sources.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("🧪 返回视频源测试页面 test_sources.html [IP:%s]", utils.GetRequestIP(r))
}

func testSourcesManageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/test_sources_manage" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := os.ReadFile("test_sources_manage.html")
	if err != nil {
		log.Printf("❌ 读取 test_sources_manage.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("🧪 返回视频源管理测试页面 test_sources_manage.html [IP:%s]", utils.GetRequestIP(r))
}

func demoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/demo" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := os.ReadFile("test_sources_manage_demo.html")
	if err != nil {
		log.Printf("❌ 读取 test_sources_manage_demo.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("🎬 返回视频源管理演示页面 test_sources_manage_demo.html [IP:%s]", utils.GetRequestIP(r))
}

func sourcesManagePreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/sources_manage_preview" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := os.ReadFile("html/sources_manage_preview.html")
	if err != nil {
		log.Printf("❌ 读取 html/sources_manage_preview.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("🎨 返回视频源管理美化预览页面 html/sources_manage_preview.html [IP:%s]", utils.GetRequestIP(r))
}

func manageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/manage" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content, err := os.ReadFile("html/manage.html")
	if err != nil {
		log.Printf("❌ 读取 html/manage.html 失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
	log.Printf("🎛️ 返回统一管理控制台页面 html/manage.html [IP:%s]", utils.GetRequestIP(r))
}

func filterConfigHandler(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 只允许GET请求
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 返回过滤配置
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"admin_password":       GlobalConfig.Filter.AdminPassword,
			"default_adult_filter": GlobalConfig.Filter.DefaultAdultFilter,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/filter_config 请求 [IP:%s]", utils.GetRequestIP(r))
}

// checkAndKillPortProcess 检查端口是否被占用，如果被占用则杀死相关进程
func checkAndKillPortProcess(port string) error {
	// 尝试监听端口来检查是否被占用
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("⚠️  端口 %s 被占用，正在查找并杀死相关进程...", port)

		// 根据操作系统查找并杀死占用端口的进程
		switch runtime.GOOS {
		case "darwin", "linux":
			return killProcessOnPortUnix(port)
		case "windows":
			return killProcessOnPortWindows(port)
		default:
			return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
		}
	}

	// 端口可用，关闭监听器
	listener.Close()
	log.Printf("✅ 端口 %s 可用", port)
	return nil
}

// killProcessOnPortUnix 在Unix系统上杀死占用指定端口的进程
func killProcessOnPortUnix(port string) error {
	cmd := exec.Command("lsof", "-ti", ":"+port)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️  未找到占用端口 %s 的进程，可能端口被系统保留", port)
		return nil
	}
	pids := strings.Fields(string(output))
	if len(pids) == 0 {
		log.Printf("✅ 端口 %s 已释放", port)
		return nil
	}
	for _, pid := range pids {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		log.Printf("🔫 正在杀死进程 PID: %s", pid)
		killCmd := exec.Command("kill", "-9", pid)
		if err := killCmd.Run(); err != nil {
			log.Printf("⚠️  杀死进程 %s 失败: %v", pid, err)
		} else {
			log.Printf("✅ 成功杀死进程 PID: %s", pid)
		}
	}
	time.Sleep(500 * time.Millisecond)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("端口 %s 仍然被占用，无法启动服务", port)
	}
	listener.Close()
	log.Printf("✅ 端口 %s 已成功释放并可用", port)
	return nil
}

// killProcessOnPortWindows 在Windows系统上杀死占用指定端口的进程
func killProcessOnPortWindows(port string) error {
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("执行 netstat 失败: %v", err)
	}
	lines := strings.Split(string(output), "\n")
	var pids []string
	for _, line := range lines {
		if strings.Contains(line, ":"+port) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pid := fields[len(fields)-1]
				if pid != "0" {
					pids = append(pids, pid)
				}
			}
		}
	}
	if len(pids) == 0 {
		log.Printf("✅ 端口 %s 已释放", port)
		return nil
	}
	for _, pid := range pids {
		log.Printf("🔫 正在杀死进程 PID: %s", pid)
		killCmd := exec.Command("taskkill", "/F", "/PID", pid)
		if err := killCmd.Run(); err != nil {
			log.Printf("⚠️  杀死进程 %s 失败: %v", pid, err)
		} else {
			log.Printf("✅ 成功杀死进程 PID: %s", pid)
		}
	}
	time.Sleep(500 * time.Millisecond)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("端口 %s 仍然被占用，无法启动服务", port)
	}
	listener.Close()
	log.Printf("✅ 端口 %s 已成功释放并可用", port)
	return nil
}
