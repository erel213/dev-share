import { useCallback, useRef, useState } from 'react'
import { Upload, X } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const ALLOWED_EXTENSIONS = ['.tf', '.tfvars', '.hcl', '.json']
const MAX_FILE_SIZE = 1024 * 1024 // 1MB

interface FileDropzoneProps {
  files: File[]
  onFilesChange: (files: File[]) => void
  error?: string
}

function validateFile(file: File): string | null {
  const ext = file.name.slice(file.name.lastIndexOf('.'))
  if (!ALLOWED_EXTENSIONS.includes(ext.toLowerCase())) {
    return `"${file.name}" has an unsupported extension. Allowed: ${ALLOWED_EXTENSIONS.join(', ')}`
  }
  if (file.size > MAX_FILE_SIZE) {
    return `"${file.name}" exceeds 1MB limit`
  }
  if (/[/\\]|\.\./.test(file.name)) {
    return `"${file.name}" contains invalid characters`
  }
  return null
}

export default function FileDropzone({
  files,
  onFilesChange,
  error,
}: FileDropzoneProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)
  const [validationError, setValidationError] = useState('')

  const addFiles = useCallback(
    (incoming: FileList | File[]) => {
      setValidationError('')
      const newFiles: File[] = []
      for (const file of Array.from(incoming)) {
        const err = validateFile(file)
        if (err) {
          setValidationError(err)
          return
        }
        if (!files.some((f) => f.name === file.name)) {
          newFiles.push(file)
        }
      }
      onFilesChange([...files, ...newFiles])
    },
    [files, onFilesChange],
  )

  const removeFile = (name: string) => {
    onFilesChange(files.filter((f) => f.name !== name))
  }

  return (
    <div className="space-y-2">
      <div
        className={cn(
          'flex cursor-pointer flex-col items-center justify-center rounded-md border-2 border-dashed p-6 transition-colors',
          dragOver
            ? 'border-primary bg-primary/5'
            : 'border-muted-foreground/25 hover:border-muted-foreground/50',
        )}
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault()
          setDragOver(true)
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(e) => {
          e.preventDefault()
          setDragOver(false)
          addFiles(e.dataTransfer.files)
        }}
      >
        <Upload className="text-muted-foreground mb-2 h-8 w-8" />
        <p className="text-muted-foreground text-sm">
          Drag & drop files here or click to browse
        </p>
        <p className="text-muted-foreground/70 mt-1 text-xs">
          {ALLOWED_EXTENSIONS.join(', ')} — max 1MB each
        </p>
      </div>
      <input
        ref={inputRef}
        type="file"
        multiple
        accept={ALLOWED_EXTENSIONS.join(',')}
        className="hidden"
        onChange={(e) => {
          if (e.target.files) addFiles(e.target.files)
          e.target.value = ''
        }}
      />
      {files.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {files.map((file) => (
            <Badge key={file.name} variant="secondary" className="gap-1 pr-1">
              {file.name}
              <button
                type="button"
                onClick={() => removeFile(file.name)}
                className="hover:bg-muted rounded-sm p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
      {(validationError || error) && (
        <p className="text-sm text-destructive">{validationError || error}</p>
      )}
    </div>
  )
}
