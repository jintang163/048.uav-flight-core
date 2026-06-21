import { useState, useEffect, useRef, useCallback } from 'react'
import { sendCommand } from '@/websocket/command'
import { speakAlert, speak } from '@/utils'

interface VoiceCommand {
  keywords: string[]
  action: string
  label: string
  requiresConfirmation: boolean
  confirmKeywords: string[]
  cancelKeywords: string[]
  execute: (uavId: string) => Promise<void>
  feedback: string
  confirmPrompt: string
}

interface VoiceLog {
  id: string
  text: string
  command: string | null
  action: string | null
  timestamp: number
  status: 'recognized' | 'executing' | 'executed' | 'cancelled' | 'rejected'
}

interface PendingConfirmation {
  command: VoiceCommand
  uavId: string
  text: string
  timestamp: number
}

interface UseVoiceControlReturn {
  isListening: boolean
  isSupported: boolean
  lastText: string
  logs: VoiceLog[]
  pendingConfirmation: PendingConfirmation | null
  startListening: () => void
  stopListening: () => void
  toggleListening: () => void
  confirmPending: () => void
  cancelPending: () => void
  clearLogs: () => void
}

const voiceCommands: VoiceCommand[] = [
  {
    keywords: ['起飞', '起飞吧', '开始起飞', '升空', '起飞', '离地'],
    action: 'takeoff',
    label: '起飞',
    requiresConfirmation: false,
    confirmKeywords: ['确认', '确定', '好', '好的', '执行'],
    cancelKeywords: ['取消', '停止', '不', '不要'],
    execute: async (uavId) => {
      await sendCommand(uavId, 'takeoff', { altitude: 5 })
    },
    feedback: '起飞指令已发送',
    confirmPrompt: '请确认起飞，说确认或取消'
  },
  {
    keywords: ['降落', '着陆', '落地', '下来', '降下来'],
    action: 'land',
    label: '降落',
    requiresConfirmation: true,
    confirmKeywords: ['确认', '确定', '好', '好的', '执行', '确认降落', '确认着陆'],
    cancelKeywords: ['取消', '停止', '不', '不要', '取消降落'],
    execute: async (uavId) => {
      await sendCommand(uavId, 'land')
    },
    feedback: '降落指令已发送',
    confirmPrompt: '降落需要二次确认，请说确认或取消'
  },
  {
    keywords: ['返航', '回来', '回家', '返回', '飞回来'],
    action: 'rtl',
    label: '返航',
    requiresConfirmation: true,
    confirmKeywords: ['确认', '确定', '好', '好的', '执行', '确认返航'],
    cancelKeywords: ['取消', '停止', '不', '不要', '取消返航'],
    execute: async (uavId) => {
      await sendCommand(uavId, 'rtl')
    },
    feedback: '返航指令已发送',
    confirmPrompt: '返航需要二次确认，请说确认或取消'
  },
  {
    keywords: ['悬停', '停下', '暂停', '停住', '定住'],
    action: 'pause',
    label: '悬停',
    requiresConfirmation: false,
    confirmKeywords: [],
    cancelKeywords: [],
    execute: async (uavId) => {
      await sendCommand(uavId, 'pause')
    },
    feedback: '悬停指令已发送',
    confirmPrompt: ''
  },
  {
    keywords: ['继续', '恢复', '继续飞', '继续飞行', '恢复飞行'],
    action: 'resume',
    label: '继续飞行',
    requiresConfirmation: false,
    confirmKeywords: [],
    cancelKeywords: [],
    execute: async (uavId) => {
      await sendCommand(uavId, 'resume')
    },
    feedback: '继续飞行指令已发送',
    confirmPrompt: ''
  },
  {
    keywords: ['解锁', '解琐', 'arm'],
    action: 'arm',
    label: '解锁',
    requiresConfirmation: false,
    confirmKeywords: [],
    cancelKeywords: [],
    execute: async (uavId) => {
      await sendCommand(uavId, 'arm')
    },
    feedback: '解锁指令已发送',
    confirmPrompt: ''
  },
  {
    keywords: ['上锁', '锁定', 'disarm'],
    action: 'disarm',
    label: '上锁',
    requiresConfirmation: true,
    confirmKeywords: ['确认', '确定', '好', '好的', '执行'],
    cancelKeywords: ['取消', '停止', '不', '不要'],
    execute: async (uavId) => {
      await sendCommand(uavId, 'disarm')
    },
    feedback: '上锁指令已发送',
    confirmPrompt: '上锁需要二次确认，请说确认或取消'
  },
  {
    keywords: ['拍照', '拍照', '照相', '拍一张', '抓拍', '截图', '拍'],
    action: 'photo',
    label: '拍照',
    requiresConfirmation: false,
    confirmKeywords: [],
    cancelKeywords: [],
    execute: async (uavId) => {
      await sendCommand(uavId, 'camera_trigger', {})
    },
    feedback: '拍照指令已发送',
    confirmPrompt: ''
  },
  {
    keywords: ['录像', '开始录像', '录制', '记录'],
    action: 'start_recording',
    label: '开始录像',
    requiresConfirmation: false,
    confirmKeywords: [],
    cancelKeywords: [],
    execute: async (uavId) => {
      await sendCommand(uavId, 'camera_control', { action: 'start_recording' })
    },
    feedback: '录像指令已发送',
    confirmPrompt: ''
  },
  {
    keywords: ['停止录像', '结束录像', '关录像', '停录像'],
    action: 'stop_recording',
    label: '停止录像',
    requiresConfirmation: false,
    confirmKeywords: [],
    cancelKeywords: [],
    execute: async (uavId) => {
      await sendCommand(uavId, 'camera_control', { action: 'stop_recording' })
    },
    feedback: '停止录像指令已发送',
    confirmPrompt: ''
  }
]

interface SpeechRecognitionEvent {
  results: SpeechRecognitionResultList
  resultIndex: number
}

interface SpeechRecognitionErrorEvent {
  error: string
  message?: string
}

interface SpeechRecognitionInstance {
  continuous: boolean
  interimResults: boolean
  lang: string
  maxAlternatives: number
  start(): void
  stop(): void
  abort(): void
  onresult: ((event: SpeechRecognitionEvent) => void) | null
  onerror: ((event: SpeechRecognitionErrorEvent) => void) | null
  onend: (() => void) | null
  onstart: (() => void) | null
}

declare global {
  interface Window {
    SpeechRecognition: new () => SpeechRecognitionInstance
    webkitSpeechRecognition: new () => SpeechRecognitionInstance
  }
}

const matchCommand = (text: string): VoiceCommand | null => {
  const normalizedText = text.toLowerCase().trim()
  for (const cmd of voiceCommands) {
    for (const keyword of cmd.keywords) {
      if (normalizedText.includes(keyword)) {
        return cmd
      }
    }
  }
  return null
}

const isConfirmWord = (text: string): boolean => {
  const normalizedText = text.toLowerCase().trim()
  const confirmWords = ['确认', '确定', '好', '好的', '执行', '是的', '对', '没错', '可以']
  return confirmWords.some(w => normalizedText.includes(w))
}

const isCancelWord = (text: string): boolean => {
  const normalizedText = text.toLowerCase().trim()
  const cancelWords = ['取消', '停止', '不', '不要', '不行', '别', '放弃']
  return cancelWords.some(w => normalizedText.includes(w))
}

export const useVoiceControl = (uavId: string | undefined): UseVoiceControlReturn => {
  const [isListening, setIsListening] = useState(false)
  const [isSupported, setIsSupported] = useState(false)
  const [lastText, setLastText] = useState('')
  const [logs, setLogs] = useState<VoiceLog[]>([])
  const [pendingConfirmation, setPendingConfirmation] = useState<PendingConfirmation | null>(null)
  const recognitionRef = useRef<SpeechRecognitionInstance | null>(null)
  const pendingRef = useRef<PendingConfirmation | null>(null)

  useEffect(() => {
    const supported = 'SpeechRecognition' in window || 'webkitSpeechRecognition' in window
    setIsSupported(supported)
  }, [])

  useEffect(() => {
    pendingRef.current = pendingConfirmation
  }, [pendingConfirmation])

  const addLog = useCallback((text: string, command: string | null, action: string | null, status: VoiceLog['status']) => {
    const log: VoiceLog = {
      id: Date.now().toString() + Math.random().toString(36).substr(2, 5),
      text,
      command,
      action,
      timestamp: Date.now(),
      status
    }
    setLogs(prev => [log, ...prev].slice(0, 50))
  }, [])

  const executeCommand = useCallback(async (command: VoiceCommand, uavId: string, text: string) => {
    addLog(text, command.action, command.label, 'executing')
    try {
      await command.execute(uavId)
      addLog(text, command.action, command.label, 'executed')
      speakAlert(command.feedback)
    } catch (error) {
      addLog(text, command.action, command.label, 'rejected')
      speak('指令执行失败')
    }
  }, [addLog])

  const handleRecognitionResult = useCallback((event: SpeechRecognitionEvent) => {
    const result = event.results[event.resultIndex]
    if (!result || !result[0]) return

    const transcript = result[0].transcript.trim()
    if (!transcript) return

    setLastText(transcript)

    const pending = pendingRef.current
    if (pending) {
      if (isConfirmWord(transcript)) {
        addLog(transcript, null, '确认' + pending.command.label, 'executed')
        executeCommand(pending.command, pending.uavId, pending.text)
        setPendingConfirmation(null)
        speak(pending.command.feedback)
      } else if (isCancelWord(transcript)) {
        addLog(transcript, null, '取消' + pending.command.label, 'cancelled')
        setPendingConfirmation(null)
        speak('已取消')
      } else {
        addLog(transcript, null, null, 'rejected')
        speak('未识别确认或取消，请重试')
      }
      return
    }

    const command = matchCommand(transcript)
    if (command) {
      if (uavId) {
        if (command.requiresConfirmation) {
          addLog(transcript, command.action, command.label, 'recognized')
          const pendingItem: PendingConfirmation = {
            command,
            uavId,
            text: transcript,
            timestamp: Date.now()
          }
          setPendingConfirmation(pendingItem)
          pendingRef.current = pendingItem
          speak(command.confirmPrompt)
        } else {
          executeCommand(command, uavId, transcript)
        }
      } else {
        addLog(transcript, command.action, command.label, 'rejected')
        speak('请先选择无人机')
      }
    } else {
      addLog(transcript, null, null, 'recognized')
    }
  }, [uavId, addLog, executeCommand])

  const startListening = useCallback(() => {
    if (!isSupported || !uavId) return

    if (recognitionRef.current) {
      recognitionRef.current.abort()
    }

    const SpeechRecognitionCtor = window.SpeechRecognition || window.webkitSpeechRecognition
    const recognition = new SpeechRecognitionCtor()
    recognition.continuous = true
    recognition.interimResults = false
    recognition.lang = 'zh-CN'
    recognition.maxAlternatives = 1

    recognition.onresult = handleRecognitionResult

    recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
      if (event.error === 'no-speech') return
      if (event.error === 'aborted') return
      console.error('Speech recognition error:', event.error)
      if (event.error === 'not-allowed') {
        setIsListening(false)
        speak('麦克风权限被拒绝')
      }
    }

    recognition.onend = () => {
      if (isListening) {
        try {
          recognition.start()
        } catch {
          setIsListening(false)
        }
      }
    }

    recognition.onstart = () => {
      setIsListening(true)
    }

    recognitionRef.current = recognition

    try {
      recognition.start()
    } catch (error) {
      console.error('Failed to start speech recognition:', error)
    }
  }, [isSupported, uavId, isListening, handleRecognitionResult])

  const stopListening = useCallback(() => {
    if (recognitionRef.current) {
      recognitionRef.current.abort()
      recognitionRef.current = null
    }
    setIsListening(false)
    setPendingConfirmation(null)
  }, [])

  const toggleListening = useCallback(() => {
    if (isListening) {
      stopListening()
    } else {
      startListening()
    }
  }, [isListening, startListening, stopListening])

  const confirmPending = useCallback(() => {
    const pending = pendingRef.current
    if (pending) {
      executeCommand(pending.command, pending.uavId, pending.text)
      setPendingConfirmation(null)
      speakAlert(pending.command.feedback)
    }
  }, [executeCommand])

  const cancelPending = useCallback(() => {
    setPendingConfirmation(null)
    speak('已取消')
  }, [])

  const clearLogs = useCallback(() => {
    setLogs([])
  }, [])

  useEffect(() => {
    return () => {
      if (recognitionRef.current) {
        recognitionRef.current.abort()
      }
    }
  }, [])

  useEffect(() => {
    if (pendingConfirmation) {
      const timer = setTimeout(() => {
        setPendingConfirmation(null)
        speak('确认超时，已取消')
      }, 15000)
      return () => clearTimeout(timer)
    }
  }, [pendingConfirmation])

  return {
    isListening,
    isSupported,
    lastText,
    logs,
    pendingConfirmation,
    startListening,
    stopListening,
    toggleListening,
    confirmPending,
    cancelPending,
    clearLogs
  }
}

export default useVoiceControl
