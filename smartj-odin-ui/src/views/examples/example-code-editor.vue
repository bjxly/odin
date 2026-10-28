<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmCodeEditor 代码编辑器示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">基于 highlight.js 的语法高亮代码编辑器，支持 JSON / SQL / Python / Shell 等语言。</p>

      <div class="mb-4">
        <el-radio-group
          v-model="language"
          size="small"
        >
          <el-radio-button value="json">JSON</el-radio-button>
          <el-radio-button value="sql">SQL</el-radio-button>
          <el-radio-button value="python">Python</el-radio-button>
          <el-radio-button value="shell">Shell</el-radio-button>
          <el-radio-button value="yaml">YAML</el-radio-button>
          <el-radio-button value="javascript">JavaScript</el-radio-button>
        </el-radio-group>
      </div>

      <SmCodeEditor
        v-model="code"
        :language="language"
        :height="350"
        placeholder="在此输入代码..."
      />
    </div>

    <div class="sm-card mt-4">
      <h3 class="text-base font-semibold mb-3">只读模式</h3>
      <SmCodeEditor
        :model-value="code"
        :language="language"
        :height="200"
        readonly
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import SmCodeEditor from '@/components/SmCodeEditor/index.vue'

const language = ref('json')

const samples: Record<string, string> = {
  json: JSON.stringify({ name: 'SmartJ', version: '1.0.0', features: ['auth', 'tenant', 'menu'] }, null, 2),
  sql: `SELECT u.username, u.nickname, d.dept_name
FROM sys_user u
LEFT JOIN sys_dept d ON u.dept_id = d.id
WHERE u.status = 1
  AND u.create_time > '2024-01-01'
ORDER BY u.create_time DESC
LIMIT 20;`,
  python: `def fibonacci(n: int) -> list[int]:
    """Generate Fibonacci sequence."""
    if n <= 0:
        return []
    seq = [0, 1]
    while len(seq) < n:
        seq.append(seq[-1] + seq[-2])
    return seq[:n]

print(fibonacci(10))`,
  shell: `#!/bin/bash
# Deploy SmartJ backend
echo "Building project..."
mvn clean package -DskipTests

echo "Starting services..."
docker-compose up -d --build

echo "Deploy completed!"`,
  yaml: `server:
  port: 8080
spring:
  application:
    name: smartj
  datasource:
    url: jdbc:mysql://localhost:3306/smartj
    username: root
mybatis-flex:
  configuration:
    map-underscore-to-camel-case: true`,
  javascript: `// 递归获取树形结构所有叶子节点
function getLeafNodes(tree, childrenKey = 'children') {
  const leaves = [];
  function traverse(nodes) {
    nodes.forEach(node => {
      if (!node[childrenKey] || !node[childrenKey].length) {
        leaves.push(node);
      } else {
        traverse(node[childrenKey]);
      }
    });
  }
  traverse(tree);
  return leaves;
}`,
}

const code = computed({
  get: () => samples[language.value],
  set: (v) => {
    samples[language.value] = v
  },
})
</script>
