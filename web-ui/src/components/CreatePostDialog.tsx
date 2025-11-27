'use client'

import { useState } from 'react'
import { Post, createPost } from '../lib/api'

interface CreatePostDialogProps {
    onPostCreated?: (post: Post) => void
}

export function CreatePostDialog({ onPostCreated }: CreatePostDialogProps) {
    const [open, setOpen] = useState(false)
    const [content, setContent] = useState('')
    const [imageUrl, setImageUrl] = useState('')
    const [isLoading, setIsLoading] = useState(false)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!content.trim()) {
            alert('Post content is required')
            return
        }

        setIsLoading(true)
        try {
            const post = await createPost({
                content: content.trim(),
                image_url: imageUrl.trim() || undefined,
            })

            onPostCreated?.(post)
            setContent('')
            setImageUrl('')
            setOpen(false)
        } catch (error) {
            console.error('Failed to create post:', error)
            alert('Failed to create post')
        } finally {
            setIsLoading(false)
        }
    }

    if (!open) {
        return (
            <button
                onClick={() => setOpen(true)}
                className="w-full bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 font-medium"
            >
                ➕ Create Post
            </button>
        )
    }

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-lg w-full mx-4">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-xl font-bold">Create a new post</h2>
                    <button
                        onClick={() => setOpen(false)}
                        className="text-gray-500 hover:text-gray-700"
                    >
                        ✕
                    </button>
                </div>

                <form onSubmit={handleSubmit}>
                    <div className="mb-4">
                        <label htmlFor="content" className="block text-sm font-medium mb-2">
                            Content
                        </label>
                        <textarea
                            id="content"
                            placeholder="What's on your mind?"
                            value={content}
                            onChange={(e) => setContent(e.target.value)}
                            rows={4}
                            required
                            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                    </div>

                    <div className="mb-4">
                        <label htmlFor="image" className="block text-sm font-medium mb-2">
                            Image URL (optional)
                        </label>
                        <input
                            id="image"
                            type="url"
                            placeholder="https://example.com/image.jpg"
                            value={imageUrl}
                            onChange={(e) => setImageUrl(e.target.value)}
                            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                    </div>

                    <div className="flex gap-2 justify-end">
                        <button
                            type="button"
                            onClick={() => setOpen(false)}
                            className="px-4 py-2 border rounded-lg hover:bg-gray-100"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                        >
                            {isLoading ? 'Creating...' : 'Create Post'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
