<template>
  <div class="tag-input">
    <el-tag
      v-for="(value, key) in modelValue"
      :key="key"
      closable
      :disable-transitions="false"
      @close="handleRemove(key)"
      style="margin-right: 8px; margin-bottom: 8px"
    >
      {{ key }}:{{ value }}
    </el-tag>
    
    <el-popover
      v-model:visible="popoverVisible"
      placement="bottom-start"
      :width="300"
      trigger="manual"
    >
      <template #reference>
        <el-button size="small" @click="togglePopover">
          + 添加标签
        </el-button>
      </template>
      
      <div class="tag-form">
        <el-form :model="tagForm" label-width="60px" size="small">
          <el-form-item label="键">
            <el-input
              v-model="tagForm.key"
              placeholder="例如: env"
              @keyup.enter="handleAdd"
            />
          </el-form-item>
          <el-form-item label="值">
            <el-input
              v-model="tagForm.value"
              placeholder="例如: production"
              @keyup.enter="handleAdd"
            />
          </el-form-item>
          <el-form-item style="margin-bottom: 0">
            <el-button type="primary" size="small" @click="handleAdd">
              添加
            </el-button>
            <el-button size="small" @click="cancelInput">取消</el-button>
          </el-form-item>
        </el-form>
        
        <!-- 常用标签建议 -->
        <div v-if="suggestions.length > 0" class="suggestions">
          <div class="suggestions-title">常用标签：</div>
          <el-tag
            v-for="(tag, index) in suggestions"
            :key="index"
            size="small"
            style="margin-right: 8px; margin-bottom: 8px; cursor: pointer"
            @click="applySuggestion(tag)"
          >
            {{ tag.key }}:{{ tag.value }}
          </el-tag>
        </div>
      </div>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

interface Props {
  modelValue: Record<string, string>
  suggestions?: Array<{ key: string; value: string }>
}

interface Emits {
  (e: 'update:modelValue', value: Record<string, string>): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => ({}),
  suggestions: () => [
    { key: 'env', value: 'production' },
    { key: 'env', value: 'staging' },
    { key: 'env', value: 'development' },
    { key: 'region', value: 'cn-north' },
    { key: 'region', value: 'cn-south' },
    { key: 'team', value: 'ops' },
    { key: 'team', value: 'dev' }
  ]
})

const emit = defineEmits<Emits>()

const popoverVisible = ref(false)
const tagForm = reactive({
  key: '',
  value: ''
})

const togglePopover = () => {
  popoverVisible.value = !popoverVisible.value
  if (popoverVisible.value) {
    // 弹出框打开时，清空表单
    tagForm.key = ''
    tagForm.value = ''
  }
}

const handleAdd = () => {
  const key = tagForm.key.trim()
  const value = tagForm.value.trim()
  
  if (!key || !value) {
    ElMessage.warning('请输入标签键和值')
    return
  }
  
  // 检查键名是否已存在
  if (props.modelValue[key]) {
    ElMessage.warning(`标签 "${key}" 已存在`)
    return
  }
  
  // 添加新标签
  const newTags = { ...props.modelValue, [key]: value }
  emit('update:modelValue', newTags)
  
  // 重置表单
  tagForm.key = ''
  tagForm.value = ''
  popoverVisible.value = false
}

const handleRemove = (key: string) => {
  const newTags = { ...props.modelValue }
  delete newTags[key]
  emit('update:modelValue', newTags)
}

const cancelInput = () => {
  tagForm.key = ''
  tagForm.value = ''
  popoverVisible.value = false
}

const applySuggestion = (tag: { key: string; value: string }) => {
  // 检查键名是否已存在
  if (props.modelValue[tag.key]) {
    ElMessage.warning(`标签 "${tag.key}" 已存在`)
    return
  }
  
  const newTags = { ...props.modelValue, [tag.key]: tag.value }
  emit('update:modelValue', newTags)
  popoverVisible.value = false
}
</script>

<style scoped lang="scss">
.tag-input {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
}

.tag-form {
  .suggestions {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #eee;
    
    .suggestions-title {
      font-size: 12px;
      color: #666;
      margin-bottom: 8px;
    }
  }
}
</style>

