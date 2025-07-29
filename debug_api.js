// 调试API问题
const API_BASE = 'http://localhost:8228';

async function debugAPI() {
    console.log('🔍 调试API问题...\n');
    
    const tests = [
        {
            name: 'GET /api/sources_manage (获取列表)',
            url: '/api/sources_manage',
            method: 'GET'
        },
        {
            name: 'PUT /api/sources_manage/jisu (更新状态)',
            url: '/api/sources_manage/jisu',
            method: 'PUT',
            body: {
                code: 'jisu',
                name: '极速资源',
                url: 'https://jisuapi.com/api.php/provide/vod',
                enabled: true,
                is_default: false
            }
        },
        {
            name: 'DELETE /api/sources_manage/jisu (删除)',
            url: '/api/sources_manage/jisu',
            method: 'DELETE'
        },
        {
            name: 'GET /api/sources_manage/test_remote (测试远程)',
            url: '/api/sources_manage/test_remote?url=https://example.com/sources.json',
            method: 'GET'
        }
    ];
    
    for (const test of tests) {
        try {
            console.log(`🧪 ${test.name}`);
            console.log(`📡 ${test.method} ${test.url}`);
            
            const options = {
                method: test.method,
                headers: {
                    'Content-Type': 'application/json'
                }
            };
            
            if (test.body) {
                options.body = JSON.stringify(test.body);
                console.log(`📦 请求体:`, JSON.stringify(test.body, null, 2));
            }
            
            const startTime = Date.now();
            const response = await fetch(`${API_BASE}${test.url}`, options);
            const endTime = Date.now();
            
            console.log(`📊 状态码: ${response.status} ${response.statusText}`);
            console.log(`⏱️ 响应时间: ${endTime - startTime}ms`);
            console.log(`📋 响应头:`, Object.fromEntries(response.headers.entries()));
            
            if (response.ok) {
                try {
                    const data = await response.json();
                    console.log(`✅ 成功: ${data.success ? '是' : '否'}`);
                    if (data.message) {
                        console.log(`💬 消息: ${data.message}`);
                    }
                    if (data.count !== undefined) {
                        console.log(`📊 数量: ${data.count}`);
                    }
                } catch (e) {
                    const text = await response.text();
                    console.log(`📄 响应文本: ${text.substring(0, 200)}...`);
                }
            } else {
                const errorText = await response.text();
                console.log(`❌ 错误响应: ${errorText}`);
            }
            
            console.log('---\n');
        } catch (error) {
            console.log(`❌ 网络错误: ${error.message}`);
            console.log('---\n');
        }
    }
    
    console.log('🏁 调试完成');
}

// 运行调试
if (typeof window === 'undefined') {
    // Node.js 环境
    debugAPI();
} else {
    // 浏览器环境
    debugAPI();
} 