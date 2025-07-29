// 调试类型标签渲染问题
const API_BASE = 'http://localhost:8228';

async function debugTypeTags() {
    console.log('🔍 调试类型标签渲染问题...\n');
    
    try {
        // 1. 获取当前源列表
        console.log('1️⃣ 获取当前源列表...');
        const sourcesResponse = await fetch(`${API_BASE}/api/sources_manage`);
        
        if (!sourcesResponse.ok) {
            console.log(`❌ 获取源列表失败: ${sourcesResponse.status}`);
            return;
        }
        
        const sourcesData = await sourcesResponse.json();
        const sources = sourcesData.sources && sourcesData.sources[0];
        
        if (!sources) {
            console.log('❌ 没有找到任何源');
            return;
        }
        
        console.log(`✅ 测试源: ${sources.name} (${sources.code})`);
        
        // 2. 测试获取类型列表
        console.log('\n2️⃣ 测试获取类型列表...');
        const typesResponse = await fetch(`${API_BASE}/api/sources_manage/types?source=${sources.code}`);
        
        if (typesResponse.ok) {
            const typesData = await typesResponse.json();
            console.log(`✅ 类型列表获取成功`);
            console.log('📊 响应数据:');
            console.log('   success:', typesData.success);
            console.log('   source:', typesData.source);
            console.log('   count:', typesData.count);
            console.log('   data类型:', Array.isArray(typesData.data) ? '数组' : typeof typesData.data);
            console.log('   data长度:', typesData.data ? typesData.data.length : 0);
            
            if (typesData.data && typesData.data.length > 0) {
                console.log('\n📝 类型数据示例:');
                const sampleTypes = typesData.data.slice(0, 3);
                sampleTypes.forEach((type, index) => {
                    console.log(`   ${index + 1}. type_id: ${type.type_id}, type_name: ${type.type_name}`);
                });
                
                // 3. 模拟渲染过程
                console.log('\n3️⃣ 模拟渲染过程...');
                console.log('🔧 渲染步骤:');
                console.log('   1. 清空容器');
                console.log('   2. 添加"全部"标签');
                console.log('   3. 添加各个类型标签');
                console.log('   4. 更新显示状态');
                
                // 4. 检查HTML结构
                console.log('\n4️⃣ 检查HTML结构...');
                console.log('🏗️ 需要的HTML元素:');
                console.log('   ✅ #typeTagsContainer - 主容器');
                console.log('   ✅ #typeCount - 类型计数');
                console.log('   ✅ #toggleTypeBtn - 切换按钮');
                console.log('   ✅ #typeSelector - 选择器容器');
                console.log('   ✅ #typeTags - 类型标签容器');
                
                // 5. 检查CSS类
                console.log('\n5️⃣ 检查CSS类...');
                console.log('🎨 需要的CSS类:');
                console.log('   ✅ .type-selector-collapsed - 收缩容器');
                console.log('   ✅ .type-selector-collapsed.expanded - 展开状态');
                console.log('   ✅ .type-tag - 类型标签');
                console.log('   ✅ .type-tag.active - 激活状态');
                console.log('   ✅ .type-tag.all-types - 全部标签');
                console.log('   ✅ .current-selection - 当前选择显示');
                
                // 6. 检查JavaScript函数
                console.log('\n6️⃣ 检查JavaScript函数...');
                console.log('⚡ 需要的JavaScript函数:');
                console.log('   ✅ loadSourceTypes() - 加载类型列表');
                console.log('   ✅ renderTypeTags() - 渲染类型标签');
                console.log('   ✅ updateSelectedTypeDisplay() - 更新显示');
                console.log('   ✅ toggleTypeSelector() - 切换展开/收起');
                console.log('   ✅ selectType() - 选择类型');
                
                // 7. 检查可能的问题
                console.log('\n7️⃣ 检查可能的问题...');
                console.log('🚨 可能的问题:');
                console.log('   1. updateSelectedTypeDisplay() 覆盖了类型标签');
                console.log('   2. 类型标签容器被隐藏');
                console.log('   3. CSS样式问题导致不可见');
                console.log('   4. JavaScript错误阻止渲染');
                console.log('   5. DOM元素不存在');
                
                // 8. 提供解决方案
                console.log('\n8️⃣ 解决方案...');
                console.log('💡 修复建议:');
                console.log('   1. 修改 updateSelectedTypeDisplay() 避免覆盖类型标签');
                console.log('   2. 确保类型标签容器在展开时可见');
                console.log('   3. 添加调试日志跟踪渲染过程');
                console.log('   4. 检查CSS样式是否正确应用');
                console.log('   5. 验证DOM元素是否正确创建');
                
            } else {
                console.log('⚠️ 类型数据为空');
            }
            
        } else {
            console.log(`❌ 类型列表请求失败: ${typesResponse.status}`);
        }
        
    } catch (error) {
        console.log(`❌ 调试错误: ${error.message}`);
    }
    
    console.log('\n🏁 类型标签调试完成');
    console.log('\n🔧 调试步骤:');
    console.log('   1. 检查浏览器控制台是否有JavaScript错误');
    console.log('   2. 检查网络请求是否成功');
    console.log('   3. 检查DOM元素是否正确创建');
    console.log('   4. 检查CSS样式是否正确应用');
    console.log('   5. 检查JavaScript函数是否正确执行');
}

// 运行调试
if (typeof window === 'undefined') {
    // Node.js 环境
    debugTypeTags();
} else {
    // 浏览器环境
    debugTypeTags();
} 