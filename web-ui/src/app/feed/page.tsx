'use client'

import { useEffect, useState } from 'react'
import { PostCard } from '../../components/PostCard'
import { CreatePostDialog } from '../../components/CreatePostDialog'
import { Layout } from '../../components/Layout'
import { Post, getPosts, getCurrentUser } from '../../lib/api'

export default function Feed() {
    const [posts, setPosts] = useState<Post[]>([])
    const [isLoading, setIsLoading] = useState(true)
    const [currentUserId, setCurrentUserId] = useState<number>()

    useEffect(() => {
        loadData()
    }, [])

    const loadData = async () => {
        try {
            const [postsData, userData] = await Promise.all([
                getPosts(),
                getCurrentUser(),
            ])
            setPosts(postsData || [])
            setCurrentUserId(userData.id)
        } catch (error) {
            console.error('Failed to load posts:', error)
        } finally {
            setIsLoading(false)
        }
    }

    const handlePostCreated = (newPost: Post) => {
        setPosts([newPost, ...posts])
    }

    const handlePostDeleted = (postId: number) => {
        setPosts(posts.filter((p) => p.id !== postId))
    }

    if (isLoading) {
        return (
            <Layout>
                <div className="flex items-center justify-center min-h-[calc(100vh-4rem)]">
                    <div className="animate-spin h-8 w-8 border-4 border-blue-600 rounded-full border-t-transparent" />
                </div>
            </Layout>
        )
    }

    return (
        <Layout>
            <div className="container max-w-2xl mx-auto py-8 px-4 space-y-6">
                <div className="sticky top-16 z-10 bg-gray-50 pb-4 pt-4">
                    <CreatePostDialog onPostCreated={handlePostCreated} />
                </div>

                <div className="space-y-6">
                    {posts.length === 0 ? (
                        <div className="text-center py-12 bg-white rounded-lg shadow">
                            <p className="text-gray-500">
                                No posts yet. Be the first to post!
                            </p>
                        </div>
                    ) : (
                        posts.map((post) => (
                            <PostCard
                                key={post.id}
                                post={post}
                                currentUserId={currentUserId}
                                onDelete={handlePostDeleted}
                            />
                        ))
                    )}
                </div>
            </div>
        </Layout>
    )
}
