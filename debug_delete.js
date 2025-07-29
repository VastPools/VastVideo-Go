// 调试删除失败问题
const API_BASE = 'http://localhost:8228';

async function debugDelete() {
    console.log('🔍 调试删除失败问题...\n');
    
    const testSource = {
        code: 'debug_test',
        name: '调试测试源',
        url: 'https://debug-test.com/api',
        enabled: true,
        is_default: false
    };
    
    try {
        // 1. 添加源
        console.log('1️⃣ 添加调试测试源...');
        const addResponse = await fetch(`${API_BASE}/api/sources_manage`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(testSource)
        });
        
        console.log(`📊 添加状态: ${addResponse.status}`);
        if (addResponse.ok) {
            const addData = await addResponse.json();
            console.log(`✅ 添加成功: ${addData.message}`);
        } else {
            const errorText = await addResponse.text();
            console.log(`❌ 添加失败: ${errorText}`);
            return;
        }
        
        // 等待一秒确保配置保存
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // 2. 检查源是否存在
        console.log('\n2️⃣ 检查源是否存在...');
        const getResponse = await fetch(`${API_BASE}/api/sources_manage/${testSource.code}`);
        console.log(`📊 获取状态: ${getResponse.status}`);
        if (getResponse.ok) {
            const getData = await getResponse.json();
            console.log(`✅ 源存在:`, getData.data);
        } else {
            const errorText = await getResponse.text();
            console.log(`❌ 源不存在: ${errorText}`);
        }
        
        // 3. 获取所有源列表
        console.log('\n3️⃣ 获取所有源列表...');
        const listResponse = await fetch(`${API_BASE}/api/sources_manage`);
        console.log(`📊 列表状态: ${listResponse.status}`);
        if (listResponse.ok) {
            const listData = await listResponse.json();
            console.log(`📊 总源数: ${listData.count}`);
            const foundSource = listData.data.find(s => s.code === testSource.code);
            if (foundSource) {
                console.log(`✅ 在列表中找到源:`, foundSource);
            } else {
                console.log(`❌ 在列表中未找到源: ${testSource.code}`);
                console.log(`📋 所有源代码:`, listData.data.map(s => s.code));
            }
        } else {
            const errorText = await listResponse.text();
            console.log(`❌ 获取列表失败: ${errorText}`);
        }
        
        // 4. 更新源
        console.log('\n4️⃣ 更新调试测试源...');
        const updateData = {
            ...testSource,
            name: '调试测试源(已更新)',
            enabled: false,
            is_default: true
        };
        
        const updateResponse = await fetch(`${API_BASE}/api/sources_manage/${testSource.code}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(updateData)
        });
        
        console.log(`📊 更新状态: ${updateResponse.status}`);
        if (updateResponse.ok) {
            const updateResult = await updateResponse.json();
            console.log(`✅ 更新成功: ${updateResult.message}`);
            console.log(`📋 更新后的数据:`, updateResult.data);
        } else {
            const errorText = await updateResponse.text();
            console.log(`❌ 更新失败: ${errorText}`);
        }
        
        // 等待一秒确保配置保存
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // 5. 再次检查源是否存在
        console.log('\n5️⃣ 更新后检查源是否存在...');
        const getResponse2 = await fetch(`${API_BASE}/api/sources_manage/${testSource.code}`);
        console.log(`📊 获取状态: ${getResponse2.status}`);
        if (getResponse2.ok) {
            const getData2 = await getResponse2.json();
            console.log(`✅ 源存在:`, getData2.data);
        } else {
            const errorText = await getResponse2.text();
            console.log(`❌ 源不存在: ${errorText}`);
        }
        
        // 6. 删除源
        console.log('\n6️⃣ 删除调试测试源...');
        const deleteResponse = await fetch(`${API_BASE}/api/sources_manage/${testSource.code}`, {
            method: 'DELETE'
        });
        
        console.log(`📊 删除状态: ${deleteResponse.status}`);
        if (deleteResponse.ok) {
            const deleteResult = await deleteResponse.json();
            console.log(`✅ 删除成功: ${deleteResult.message}`);
        } else {
            const errorText = await deleteResponse.text();
            console.log(`❌ 删除失败: ${errorText}`);
        }
        
    } catch (error) {
        console.log(`❌ 网络错误: ${error.message}`);
    }
    
    console.log('\n🏁 调试完成');
}

// 运行调试
if (typeof window === 'undefined') {
    // Node.js 环境
    debugDelete();
} else {
    // 浏览器环境
    debugDelete();
} 