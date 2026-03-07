import { useCallback, useRef, useState } from 'react'
import { Upload, X, FolderUp } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const ALLOWED_EXTENSIONS = ['.tf', '.tfvars', '.hcl', '.json']
const MAX_FILE_SIZE = 1024 * 1024 // 1MB

export interface FileWithPath {
  file: File
  path: string // e.g., "modules/vpc/main.tf"
}

interface FileDropzoneProps {
  files: FileWithPath[]
  onFilesChange: (files: FileWithPath[]) => void
  error?: string
}

function hasAllowedExtension(path: string): boolean {
  const ext = path.slice(path.lastIndexOf('.'))
  return ALLOWED_EXTENSIONS.includes(ext.toLowerCase())
}

function validateFilePath(path: string): string | null {
  if (/\\|\.\./.test(path)) {
    return `"${path}" contains invalid characters`
  }
  return null
}

function validateFileSize(file: File, path: string): string | null {
  if (file.size > MAX_FILE_SIZE) {
    return `"${path}" exceeds 1MB limit`
  }
  return null
}

interface FileSystemEntry {
  isFile: boolean
  isDirectory: boolean
  name: string
  file(cb: (file: File) => void): void
  createReader(): { readEntries(cb: (entries: FileSystemEntry[]) => void): void }
  fullPath: string
}

async function traverseEntry(entry: FileSystemEntry, basePath: string): Promise<FileWithPath[]> {
  if (entry.isFile) {
    return new Promise((resolve) => {
      entry.file((file) => {
        const path = basePath ? `${basePath}/${entry.name}` : entry.name
        resolve([{ file, path }])
      })
    })
  }
  if (entry.isDirectory) {
    return new Promise((resolve) => {
      const reader = entry.createReader()
      reader.readEntries(async (entries) => {
        const newBase = basePath ? `${basePath}/${entry.name}` : entry.name
        const results: FileWithPath[] = []
        for (const child of entries) {
          const childFiles = await traverseEntry(child, newBase)
          results.push(...childFiles)
        }
        resolve(results)
      })
    })
  }
  return []
}

function stripTopLevelDir(files: FileWithPath[]): FileWithPath[] {
  if (files.length === 0) return files
  const firstSlash = files[0].path.indexOf('/')
  if (firstSlash === -1) return files // no directory prefix
  const prefix = files[0].path.slice(0, firstSlash + 1)
  const allSamePrefix = files.every((f) => f.path.startsWith(prefix))
  if (!allSamePrefix) return files
  return files.map((f) => ({ ...f, path: f.path.slice(prefix.length) }))
}

export default function FileDropzone({
  files,
  onFilesChange,
  error,
}: FileDropzoneProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const folderInputRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)
  const [validationError, setValidationError] = useState('')

  const addFiles = useCallback(
    (incoming: FileWithPath[]) => {
      setValidationError('')
      const newFiles: FileWithPath[] = []
      for (const item of incoming) {
        if (!hasAllowedExtension(item.path)) continue
        const pathErr = validateFilePath(item.path)
        if (pathErr) {
          setValidationError(pathErr)
          return
        }
        const sizeErr = validateFileSize(item.file, item.path)
        if (sizeErr) {
          setValidationError(sizeErr)
          return
        }
        if (!files.some((f) => f.path === item.path)) {
          newFiles.push(item)
        }
      }
      const merged = [...files, ...newFiles]
      console.log('[FileDropzone] addFiles result:', merged.map((f) => ({ path: f.path, fileName: f.file.name, webkitRelativePath: f.file.webkitRelativePath })))
      onFilesChange(merged)
    },
    [files, onFilesChange],
  )

  const handleFileInput = useCallback(
    (fileList: FileList | null) => {
      if (!fileList) return
      const items: FileWithPath[] = Array.from(fileList).map((file) => ({
        file,
        path: file.name,
      }))
      addFiles(items)
    },
    [addFiles],
  )

  const handleFolderInput = useCallback(
    (fileList: FileList | null) => {
      if (!fileList) return
      const items: FileWithPath[] = Array.from(fileList).map((file) => {
        // webkitRelativePath is like "folderName/sub/file.tf", strip top-level dir
        const relPath = file.webkitRelativePath
        const firstSlash = relPath.indexOf('/')
        const path = firstSlash !== -1 ? relPath.slice(firstSlash + 1) : relPath
        return { file, path }
      })
      addFiles(items)
    },
    [addFiles],
  )

  const handleDrop = useCallback(
    async (e: React.DragEvent) => {
      e.preventDefault()
      setDragOver(false)

      const items = e.dataTransfer.items
      if (items && items.length > 0) {
        const entries: FileSystemEntry[] = []
        for (let i = 0; i < items.length; i++) {
          const entry = (items[i] as unknown as { webkitGetAsEntry(): FileSystemEntry | null }).webkitGetAsEntry?.()
          if (entry) entries.push(entry)
        }

        if (entries.length > 0) {
          const allFiles: FileWithPath[] = []
          for (const entry of entries) {
            const result = await traverseEntry(entry, '')
            allFiles.push(...result)
          }
          // If a single directory was dropped, strip its name from all paths
          if (entries.length === 1 && entries[0].isDirectory) {
            addFiles(stripTopLevelDir(allFiles))
          } else {
            addFiles(allFiles)
          }
          return
        }
      }

      // Fallback for browsers without webkitGetAsEntry
      const fileItems: FileWithPath[] = Array.from(e.dataTransfer.files).map(
        (file) => ({ file, path: file.name }),
      )
      addFiles(fileItems)
    },
    [addFiles],
  )

  const removeFile = (path: string) => {
    onFilesChange(files.filter((f) => f.path !== path))
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
        onDrop={handleDrop}
      >
        <Upload className="text-muted-foreground mb-2 h-8 w-8" />
        <p className="text-muted-foreground text-sm">
          Drag & drop files or folders here, or click to browse files
        </p>
        <p className="text-muted-foreground/70 mt-1 text-xs">
          {ALLOWED_EXTENSIONS.join(', ')} — max 1MB each
        </p>
      </div>
      <div className="flex gap-2">
        <input
          ref={inputRef}
          type="file"
          multiple
          accept={ALLOWED_EXTENSIONS.join(',')}
          className="hidden"
          onChange={(e) => {
            handleFileInput(e.target.files)
            e.target.value = ''
          }}
        />
        <button
          type="button"
          onClick={() => folderInputRef.current?.click()}
          className="inline-flex items-center gap-1 rounded-md border px-3 py-1.5 text-xs text-muted-foreground hover:bg-accent"
        >
          <FolderUp className="h-3.5 w-3.5" />
          Browse Folder
        </button>
        <input
          ref={folderInputRef}
          type="file"
          className="hidden"
          onChange={(e) => {
            handleFolderInput(e.target.files)
            e.target.value = ''
          }}
          {...({ webkitdirectory: '', directory: '' } as React.InputHTMLAttributes<HTMLInputElement>)}
        />
      </div>
      {files.length > 0 && (
        <div className="flex max-h-40 flex-wrap gap-2 overflow-y-auto">
          {files.map((f) => (
            <Badge key={f.path} variant="secondary" className="gap-1 pr-1">
              {f.path}
              <button
                type="button"
                onClick={() => removeFile(f.path)}
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
