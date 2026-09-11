import { ref } from 'vue'

export interface ConfirmOptions {
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'danger' | 'primary'
}

const isOpen = ref(false)
const options = ref<ConfirmOptions>({
  title: '',
  message: '',
  confirmLabel: 'CONFIRMER',
  cancelLabel: 'ANNULER',
  variant: 'primary',
})

let resolvePromise: ((val: boolean) => void) | null = null

export const useConfirmState = () => {
  const handleConfirm = () => {
    isOpen.value = false
    if (resolvePromise) {
      resolvePromise(true)
      resolvePromise = null
    }
  }

  const handleCancel = () => {
    isOpen.value = false
    if (resolvePromise) {
      resolvePromise(false)
      resolvePromise = null
    }
  }

  return {
    isOpen,
    options,
    handleConfirm,
    handleCancel,
  }
}

export const useConfirm = () => {
  const confirm = (opts: ConfirmOptions): Promise<boolean> => {
    options.value = {
      confirmLabel: 'CONFIRMER',
      cancelLabel: 'ANNULER',
      variant: 'primary',
      ...opts,
    }
    isOpen.value = true

    return new Promise<boolean>((resolve) => {
      resolvePromise = resolve
    })
  }

  return {
    confirm,
  }
}
