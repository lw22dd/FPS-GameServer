const http = require('http');

console.log('=== 测试游戏服务器 ===\n');

async function runTests() {
  try {
    console.log('1. 测试服务器状态...');
    await testRequest('/user/test', 'GET');
    
    console.log('\n2. 测试用户注册...');
    await testRequest('/user/register', 'POST', {
      username: 'player1',
      password: '123456',
      email: 'player1@test.com'
    });

    console.log('\n3. 测试用户名重复...');
    await testRequest('/user/register', 'POST', {
      username: 'player1',
      password: '123456',
      email: 'player2@test.com'
    });

    console.log('\n4. 测试用户名空格检查...');
    await testRequest('/user/register', 'POST', {
      username: 'player 1',
      password: '123456',
      email: 'player3@test.com'
    });

    console.log('\n5. 测试用户登录...');
    await testRequest('/user/login', 'POST', {
      username: 'player1',
      password: '123456'
    });

    console.log('\n6. 测试密码错误...');
    await testRequest('/user/login', 'POST', {
      username: 'player1',
      password: 'wrong'
    });

    console.log('\n7. 测试用户不存在...');
    await testRequest('/user/login', 'POST', {
      username: 'notexist',
      password: '123456'
    });

    console.log('\n8. 测试创建房间...');
    await testRequest('/rooms/create?username=player1', 'POST', {
      name: 'TestRoom',
      maxPlayers: 2
    });

    console.log('\n9. 测试获取房间列表...');
    await testRequest('/rooms/list', 'GET');

    console.log('\n=== 测试完成 ===');
    process.exit(0);
  } catch (error) {
    console.error('测试失败:', error.message);
    process.exit(1);
  }
}

function testRequest(path, method, data = null) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      const url = new URL(path, 'http://localhost:8080');
      
      const options = {
        hostname: 'localhost',
        port: 8080,
        path: url.pathname + url.search,
        method: method,
        headers: {
          'Content-Type': 'application/json'
        },
        timeout: 5000
      };

      const req = http.request(options, (res) => {
        let body = '';
        res.on('data', chunk => body += chunk);
        res.on('end', () => {
          try {
            const json = JSON.parse(body);
            console.log(`   状态: ${res.statusCode}`);
            console.log(`   响应: ${JSON.stringify(json, null, 2)}`);
            resolve();
          } catch (e) {
            console.log(`   状态: ${res.statusCode}`);
            console.log(`   响应: ${body}`);
            resolve();
          }
        });
      });

      req.on('error', (e) => {
        console.log(`   错误: ${e.message}`);
        resolve();
      });

      req.on('timeout', () => {
        console.log(`   超时`);
        req.destroy();
        resolve();
      });

      if (data) {
        req.write(JSON.stringify(data));
      }
      req.end();
    }, 500);
  });
}

runTests();
