 'use client'
 
 import { useState } from 'react'
 import { useRouter } from 'next/navigation'
 import { SquarePlus, Settings, Trash2 } from 'lucide-react'
 import { Button } from '@workspace/ui/components/atoms/button'
 import { Input } from '@workspace/ui/components/atoms/input'
 import { Badge } from '@workspace/ui/components/atoms/badge'
 import { PagesTable } from '@/features/page/ui/pages-table'
 import { AddPageDialog } from '@/features/page/ui/add-page-dialog'
 import { EditPageDialog } from '@/features/page/ui/edit-page-dialog'
 import { DeletePageDialog } from '@/features/page/ui/delete-page-dialog'
 import { PageApi } from '@workspace/services/page-api'
 import { EditWorkspaceDialog } from '@/features/workspace/ui/edit-workspace-dialog'
 import { DeleteWorkspaceDialog } from '@/features/workspace/ui/delete-workspace-dialog'
 import { useWorkspaces } from '@/features/workspace/application/hooks/use-workspaces'
 import type { Page, CreatePageDto } from '@/features/page/domain/types'
 import type { Workspace, WorkspaceType } from '@/features/workspace/domain/types'
 
 export interface WorkspaceDetailContentProps {
   workspace: Workspace
   initialPages?: Page[]
 }
 
 export function WorkspaceDetailContent({
   workspace: initialWorkspace,
   initialPages = [],
 }: Readonly<WorkspaceDetailContentProps>) {
   const router = useRouter()
   const { updateWorkspace, deleteWorkspace, isLoading: isWorkspaceLoading } = useWorkspaces()
 
   const [workspace, setWorkspace] = useState<Workspace>(initialWorkspace)
   const [pages, setPages] = useState<Page[]>(initialPages)
   const [searchQuery, setSearchQuery] = useState('')
   const [isAddPageOpen, setIsAddPageOpen] = useState(false)
   const [isEditWorkspaceOpen, setIsEditWorkspaceOpen] = useState(false)
   const [isDeleteWorkspaceOpen, setIsDeleteWorkspaceOpen] = useState(false)
   const [isLoading, setIsLoading] = useState(false)
   const [error, setError] = useState<Error | null>(null)
 
   const [isEditPageOpen, setIsEditPageOpen] = useState(false)
   const [isDeletePageOpen, setIsDeletePageOpen] = useState(false)
   const [selectedPage, setSelectedPage] = useState<Page | null>(null)
 
   const handleAddPage = async (data: CreatePageDto) => {
     setIsLoading(true)
     setError(null)
 
     try {
       const newPage = await PageApi.createPage(data)
       setPages((prev) => [
         newPage,
         ...prev,
       ])
       setIsAddPageOpen(false)
     } catch (err) {
       setError(err instanceof Error ? err : new Error('Failed to add page'))
     } finally {
       setIsLoading(false)
     }
   }
 
   const handleEditPageClick = (page: Page) => {
     setSelectedPage(page)
     setIsEditPageOpen(true)
   }
 
   const handleDeletePageClick = (page: Page) => {
     setSelectedPage(page)
     setIsDeletePageOpen(true)
   }
 
   const handleUpdatePage = async (
     pageId: string,
     data: {
       name: string
       url: string
     }
   ) => {
     setIsLoading(true)
     try {
       const updatedPage = await PageApi.updatePage(pageId, data)
       setPages((prev) => prev.map((p) => (p.id === pageId ? updatedPage : p)))
       setIsEditPageOpen(false)
       setSelectedPage(null)
     } catch (err) {
       console.error('Failed to update page:', err)
     } finally {
       setIsLoading(false)
     }
   }
 
   const handleDeletePage = async () => {
     if (!selectedPage) return
     setIsLoading(true)
     try {
       await PageApi.deletePage(selectedPage.id)
       setPages((prev) => prev.filter((p) => p.id !== selectedPage.id))
       setIsDeletePageOpen(false)
       setSelectedPage(null)
     } catch (err) {
       console.error('Failed to delete page:', err)
     } finally {
       setIsLoading(false)
     }
   }
 
   const handleUpdateWorkspace = async (
     id: string,
     data: {
       name: string
       type: WorkspaceType
       tags: string[]
     }
   ) => {
     try {
       const updated = await updateWorkspace(id, data)
       if (updated) {
         setWorkspace(updated)
         setIsEditWorkspaceOpen(false)
         router.refresh()
       }
     } catch (err) {
       console.error('Failed to update workspace:', err)
     }
   }
 
   const handleDeleteWorkspace = async () => {
     try {
       await deleteWorkspace(workspace.id)
       router.push('/workspaces')
     } catch (err) {
       console.error('Failed to delete workspace:', err)
     }
   }
 
   const handleViewChanges = (pageId: string) => {
     router.push(`/workspaces/${workspace.id}/pages/${pageId}`)
   }
 
   const handlePageClick = (pageId: string) => {
     router.push(`/workspaces/${workspace.id}/pages/${pageId}`)
   }
 
   const handleCheckFrequencyChange = async (pageId: string, frequency: string) => {
     setPages((prev) =>
       prev.map((page) =>
         page.id === pageId
           ? {
               ...page,
               checkFrequency: frequency,
             }
           : page
       )
     )
 
     try {
       await PageApi.updateMonitoringConfig(pageId, {
         checkFrequency: frequency,
       })
     } catch (err) {
       setPages(initialPages)
       console.error('Failed to update check frequency:', err)
     }
   }
 
   const filteredPages = pages.filter((page) =>
     page.name.toLowerCase().includes(searchQuery.toLowerCase())
   )
 
   return (
     <div className="flex-1 flex flex-col bg-background">
       <div className="flex justify-between items-start gap-2 px-24 py-6">
         <div className="flex flex-col gap-2">
           <div className="flex items-center gap-3">
             <h1 className="text-2xl font-semibold text-foreground">
               Added pages for {workspace.name}
             </h1>
             <div className="flex gap-1">
               {workspace.tags?.map((tag) => (
                   <Badge key={tag} variant="secondary">
                     {tag}
                   </Badge>
                 ))}
             </div>
           </div>
           <p className="text-base font-normal text-muted-foreground">
             Here are all the pages you've added to this workspace.
           </p>
         </div>
 
         <div className="flex items-center gap-2">
           <Button variant="outline" onClick={() => setIsEditWorkspaceOpen(true)} className="gap-2">
             <Settings className="w-4 h-4" />
             Edit Workspace
           </Button>
           <Button
             variant="destructive"
             onClick={() => setIsDeleteWorkspaceOpen(true)}
             size="icon"
             className="h-10 w-10"
           >
             <Trash2 className="w-4 h-4" />
           </Button>
         </div>
       </div>
 
       <div className="flex justify-between items-center px-24 py-2 gap-4">
         <div className="relative flex-1 max-w-sm">
           <svg
             width="17"
             height="17"
             viewBox="0 0 17 17"
             fill="none"
             className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
           >
             <path
               d="M7.79167 13.4583C10.8292 13.4583 13.2917 10.9958 13.2917 7.95833C13.2917 4.92084 10.8292 2.45833 7.79167 2.45833C4.75418 2.45833 2.29167 4.92084 2.29167 7.95833C2.29167 10.9958 4.75418 13.4583 7.79167 13.4583Z"
               stroke="currentColor"
               strokeWidth="1.5"
               strokeLinecap="round"
               strokeLinejoin="round"
             />
             <path
               d="M14.5833 14.75L11.7292 11.8958"
               stroke="currentColor"
               strokeWidth="1.5"
               strokeLinecap="round"
               strokeLinejoin="round"
             />
           </svg>
           <Input
             type="search"
             placeholder="Search pages"
             value={searchQuery}
             onChange={(e) => setSearchQuery(e.target.value)}
             className="pl-10"
           />
         </div>
 
         <div className="flex items-center gap-4">
           <Button
             variant="default"
             onClick={() => setIsAddPageOpen(true)}
             className="h-10 px-4 gap-2 bg-primary"
           >
             <SquarePlus className="w-4 h-4" />
             Add page
           </Button>
         </div>
       </div>
 
       <div className="px-24 py-2 pb-6">
         <PagesTable
           pages={filteredPages}
           onViewChanges={handleViewChanges}
           onPageClick={handlePageClick}
           onCheckFrequencyChange={handleCheckFrequencyChange}
           onEdit={handleEditPageClick}
           onDelete={handleDeletePageClick}
         />
       </div>
 
       <AddPageDialog
         open={isAddPageOpen}
         onOpenChange={setIsAddPageOpen}
         onSubmit={handleAddPage}
         workspaceId={workspace.id}
         isLoading={isLoading}
         error={error}
       />
 
       <EditPageDialog
         open={isEditPageOpen}
         onOpenChange={setIsEditPageOpen}
         onSubmit={handleUpdatePage}
         page={selectedPage}
         isLoading={isLoading}
       />
 
       <DeletePageDialog
         open={isDeletePageOpen}
         onOpenChange={setIsDeletePageOpen}
         onConfirm={handleDeletePage}
         pageName={selectedPage?.name ?? ''}
         isLoading={isLoading}
       />
 
       <EditWorkspaceDialog
         open={isEditWorkspaceOpen}
         onOpenChange={setIsEditWorkspaceOpen}
         onSubmit={handleUpdateWorkspace}
         isLoading={isWorkspaceLoading}
         error={null}
         workspace={workspace}
       />
 
       <DeleteWorkspaceDialog
         open={isDeleteWorkspaceOpen}
         onOpenChange={setIsDeleteWorkspaceOpen}
         onConfirm={handleDeleteWorkspace}
         workspaceName={workspace.name}
         isLoading={isWorkspaceLoading}
       />
     </div>
   )
 }
