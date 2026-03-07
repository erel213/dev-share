import { useCallback, useEffect, useMemo, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import {
  ArrowLeft,
  ChevronDown,
  ChevronRight,
  File,
  FileCode,
  FileJson,
  Folder,
  FolderOpen,
} from 'lucide-react'
import {
  getTemplate,
  getTemplateFiles,
  getTemplateFileContent,
} from '@/lib/templates-api'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import type { ApiError, Template, TemplateFileInfo } from '@/types/api'

interface TreeNode {
  name: string
  path: string
  isDir: boolean
  children: TreeNode[]
}

function buildTree(files: TemplateFileInfo[]): TreeNode[] {
  const root: TreeNode[] = []

  for (const file of files) {
    const parts = file.name.split('/')
    let current = root
    let currentPath = ''

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i]
      currentPath = currentPath ? `${currentPath}/${part}` : part
      const isLast = i === parts.length - 1

      let existing = current.find((n) => n.name === part)
      if (!existing) {
        existing = {
          name: part,
          path: currentPath,
          isDir: !isLast,
          children: [],
        }
        current.push(existing)
      }
      current = existing.children
    }
  }

  function sortNodes(nodes: TreeNode[]): TreeNode[] {
    return nodes
      .sort((a, b) => {
        if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
        return a.name.localeCompare(b.name)
      })
      .map((n) => ({ ...n, children: sortNodes(n.children) }))
  }

  return sortNodes(root)
}

function fileIcon(name: string) {
  const ext = name.split('.').pop()?.toLowerCase()
  if (ext === 'tf' || ext === 'hcl')
    return <FileCode className="h-4 w-4 text-muted-foreground" />
  if (ext === 'json')
    return <FileJson className="h-4 w-4 text-muted-foreground" />
  return <File className="h-4 w-4 text-muted-foreground" />
}

function TreeNodeView({
  node,
  depth,
  selectedFile,
  onFileClick,
}: {
  node: TreeNode
  depth: number
  selectedFile: string | null
  onFileClick: (path: string) => void
}) {
  const [expanded, setExpanded] = useState(true)

  if (node.isDir) {
    return (
      <div>
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex w-full items-center gap-1 rounded px-2 py-1 text-sm hover:bg-accent text-left"
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
        >
          {expanded ? (
            <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
          )}
          {expanded ? (
            <FolderOpen className="h-4 w-4 text-muted-foreground" />
          ) : (
            <Folder className="h-4 w-4 text-muted-foreground" />
          )}
          <span className="ml-1">{node.name}</span>
        </button>
        {expanded &&
          node.children.map((child) => (
            <TreeNodeView
              key={child.path}
              node={child}
              depth={depth + 1}
              selectedFile={selectedFile}
              onFileClick={onFileClick}
            />
          ))}
      </div>
    )
  }

  return (
    <button
      onClick={() => onFileClick(node.path)}
      className={`flex w-full items-center gap-2 rounded px-2 py-1 text-sm hover:bg-accent text-left ${
        selectedFile === node.path ? 'bg-accent text-accent-foreground' : ''
      }`}
      style={{ paddingLeft: `${depth * 16 + 24}px` }}
    >
      {fileIcon(node.name)}
      {node.name}
    </button>
  )
}

export default function TemplateBrowserPage() {
  const { id } = useParams<{ id: string }>()

  const [template, setTemplate] = useState<Template | null>(null)
  const [files, setFiles] = useState<TemplateFileInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [fileContent, setFileContent] = useState<string>('')
  const [contentLoading, setContentLoading] = useState(false)

  const tree = useMemo(() => buildTree(files), [files])

  const fetchData = useCallback(async () => {
    if (!id) return
    setError('')
    setLoading(true)
    try {
      const [tmpl, fileList] = await Promise.all([
        getTemplate(id),
        getTemplateFiles(id),
      ])
      setTemplate(tmpl)
      setFiles(fileList)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load template')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleFileClick = useCallback(
    async (filePath: string) => {
      if (!id) return
      setSelectedFile(filePath)
      setContentLoading(true)
      setFileContent('')
      try {
        const content = await getTemplateFileContent(id, filePath)
        setFileContent(content)
      } catch (err) {
        const apiError = err as ApiError
        setFileContent(
          `Error loading file: ${apiError.message || 'Unknown error'}`,
        )
      } finally {
        setContentLoading(false)
      }
    },
    [id],
  )

  if (error) {
    return (
      <div className="space-y-4">
        <Link to="/templates">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Templates
          </Button>
        </Link>
        <p className="text-sm text-destructive">{error}</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Link to="/templates">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Button>
        </Link>
        {loading ? (
          <Skeleton className="h-8 w-48" />
        ) : (
          <h1 className="text-2xl font-bold tracking-tight">
            {template?.name}
          </h1>
        )}
      </div>

      <div
        className="flex gap-4 rounded-md border"
        style={{ minHeight: '500px' }}
      >
        {/* File tree - left panel */}
        <div className="w-1/4 border-r p-3">
          {loading ? (
            <div className="space-y-2">
              <Skeleton className="h-5 w-32" />
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="ml-4 h-4 w-28" />
              ))}
            </div>
          ) : (
            <div className="space-y-0.5">
              <div className="flex items-center gap-2 font-medium text-sm py-1 px-2">
                <FolderOpen className="h-4 w-4 text-muted-foreground" />
                {template?.name}
              </div>
              {tree.map((node) => (
                <TreeNodeView
                  key={node.path}
                  node={node}
                  depth={1}
                  selectedFile={selectedFile}
                  onFileClick={handleFileClick}
                />
              ))}
              {files.length === 0 && (
                <p className="text-muted-foreground pl-6 text-sm">No files</p>
              )}
            </div>
          )}
        </div>

        {/* Content viewer - right panel */}
        <div className="flex-1 p-3 overflow-auto">
          {!selectedFile ? (
            <p className="text-muted-foreground text-sm">
              Select a file to view its content.
            </p>
          ) : contentLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 8 }).map((_, i) => (
                <Skeleton
                  key={i}
                  className="h-4"
                  style={{ width: `${60 + Math.random() * 30}%` }}
                />
              ))}
            </div>
          ) : (
            <div>
              <div className="mb-2 text-sm font-medium text-muted-foreground">
                {selectedFile}
              </div>
              <pre className="rounded-md bg-muted p-4 text-sm overflow-auto">
                <code>{fileContent}</code>
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
