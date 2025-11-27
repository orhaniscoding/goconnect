'use client'

import { useState } from 'react'
import { Post, deletePost, likePost, unlikePost } from '../lib/api'

interface PostCardProps {
    post: Post
    currentUserId?: number
    onDelete?: (postId: number) => void
    onEdit?: (post: Post) => void
}

export function PostCard({ post, currentUserId, onDelete, onEdit }: PostCardProps) {
    const [likes, setLikes] = useState(post.likes)
    const [isLiked, setIsLiked] = useState(false)
    const [isDeleting, setIsDeleting] = useState(false)

    const isOwnPost = currentUserId === post.user_id

    const formatTimeAgo = (dateString: string) => {
        const date = new Date(dateString)
        const now = new Date()
        const seconds = Math.floor((now.getTime() - date.getTime()) / 1000)

        if (seconds < 60) return `${seconds}s ago`
        const minutes = Math.floor(seconds / 60)
        if (minutes < 60) return `${minutes}m ago`
        const hours = Math.floor(minutes / 60)
        if (hours < 24) return `${hours}h ago`
        const days = Math.floor(hours / 24)
        return `${days}d ago`
    }

    const handleLike = async () => {
        try {
            if (isLiked) {
                await unlikePost(post.id)
                setLikes((prev: number) => Math.max(0, prev - 1))
            } else {
                await likePost(post.id)
                setLikes((prev: number) => prev + 1)
            }
            setIsLiked(!isLiked)
        } catch (error) {
            console.error('Failed to update like:', error)
        }
    }

    const handleDelete = async () => {
        if (!window.confirm('Are you sure you want to delete this post?')) return

        setIsDeleting(true)
        try {
            await deletePost(post.id)
            onDelete?.(post.id)
        } catch (error) {
            console.error('Failed to delete post:', error)
        } finally {
            setIsDeleting(false)
        }
    }

    return (
        <div className="bg-white rounded-lg shadow p-6 mb-4">
            <div className="flex items-center gap-4 mb-4">
                <div className="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center">
                    {post.author?.avatar_url ? (
                        <img
                            src={post.author.avatar_url}
                            alt={post.author.username}
                            className="w-12 h-12 rounded-full object-cover"
                        />
                    ) : (
                        <span className="text-xl font-bold text-gray-600">
                            {post.author?.username.charAt(0).toUpperCase()}
                        </span>
                    )}
                </div>
                <div className="flex-1">
                    <div className="flex items-center justify-between">
                        <div>
                            <p className="font-semibold">{post.author?.full_name || post.author?.username}</p>
                            <p className="text-sm text-gray-500">
                                @{post.author?.username} • {formatTimeAgo(post.created_at)}
                            </p>
                        </div>
                        {isOwnPost && (
                            <div className="relative">
                                <button className="p-2 hover:bg-gray-100 rounded">⋮</button>
                                <div className="hidden absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg">
                                    <button
                                        onClick={() => onEdit?.(post)}
                                        className="block w-full text-left px-4 py-2 hover:bg-gray-100"
                                    >
                                        Edit
                                    </button>
                                    <button
                                        onClick={handleDelete}
                                        disabled={isDeleting}
                                        className="block w-full text-left px-4 py-2 hover:bg-gray-100 text-red-600"
                                    >
                                        Delete
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            <div className="mb-4">
                <p className="whitespace-pre-wrap">{post.content}</p>
                {post.image_url && (
                    <img
                        src={post.image_url}
                        alt="Post image"
                        className="w-full rounded-lg mt-4 max-h-96 object-cover"
                    />
                )}
            </div>

            <div className="flex gap-4 pt-4 border-t">
                <button
                    onClick={handleLike}
                    className={`flex items-center gap-2 px-4 py-2 rounded hover:bg-gray-100 ${isLiked ? 'text-red-500' : ''
                        }`}
                >
                    <span>{isLiked ? '❤️' : '🤍'}</span>
                    <span>{likes}</span>
                </button>
                <button className="flex items-center gap-2 px-4 py-2 rounded hover:bg-gray-100">
                    <span>💬</span>
                    <span>Comment</span>
                </button>
                <button className="flex items-center gap-2 px-4 py-2 rounded hover:bg-gray-100">
                    <span>🔗</span>
                    <span>Share</span>
                </button>
            </div>
        </div>
    )
}
