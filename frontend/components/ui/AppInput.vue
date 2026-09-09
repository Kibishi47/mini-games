<template>
  <input
    :type="type"
    :value="modelValue"
    :placeholder="placeholder"
    :maxlength="maxlength"
    :disabled="disabled"
    :required="required"
    :class="[
      'w-full bg-board-white border-[3px] border-ink-black rounded-xl px-4 py-2.5 text-ink-black font-body font-semibold',
      'placeholder:text-ink-black/40 focus:outline-none focus:shadow-pop-sm transition-none',
      'disabled:opacity-50 disabled:bg-board-cream disabled:cursor-not-allowed',
      customClass
    ]"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue?: string | number
    type?: string
    placeholder?: string
    maxlength?: number | string
    disabled?: boolean
    required?: boolean
    customClass?: string
  }>(),
  {
    modelValue: '',
    type: 'text',
    placeholder: '',
    disabled: false,
    required: false,
    customClass: '',
  }
)

defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()
</script>
