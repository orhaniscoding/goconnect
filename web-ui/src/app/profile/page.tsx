'use client'

import { useEffect, useState } from 'react'
import { Layout } from '../../components/Layout'
import { getCurrentUser, updateProfile } from '../../lib/api'

export default function Profile() {
    const [user, setUser] = useState<any>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [isEditing, setIsEditing] = useState(false)
    const [isSaving, setIsSaving] = useState(false)
    const [formData, setFormData] = useState({
        full_name: '',
        bio: '',
        avatar_url: '',
    })

    useEffect(() => {
        loadProfile()
    }, [])

    const loadProfile = async () => {
        try {
            const userData = await getCurrentUser()
            setUser(userData)
            setFormData({
                full_name: userData.full_name || '',
                bio: userData.bio || '',
                avatar_url: userData.avatar_url || '',
            })
        } catch (error) {
            console.error('Failed to load profile:', error)
        } finally {
            setIsLoading(false)
        }
    }

    const handleSave = async () => {
        setIsSaving(true)
        try {
            await updateProfile({
                full_name: formData.full_name || undefined,
                bio: formData.bio || undefined,
                avatar_url: formData.avatar_url || undefined,
            })
            await loadProfile()
            setIsEditing(false)
            alert('Profile updated successfully')
        } catch (error) {
            console.error('Failed to update profile:', error)
            alert('Failed to update profile')
        } finally {
            setIsSaving(false)
        }
    }

    const handleCancel = () => {
        setFormData({
            full_name: user?.full_name || '',
            bio: user?.bio || '',
            avatar_url: user?.avatar_url || '',
        })
        setIsEditing(false)
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
            <div className="container max-w-2xl mx-auto py-8 px-4">
                <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex items-center justify-between mb-6">
                        <h2 className="text-2xl font-bold">Profile</h2>
                        {!isEditing ? (
                            <button
                                onClick={() => setIsEditing(true)}
                                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                            >
                                ✏️ Edit Profile
                            </button>
                        ) : (
                            <div className="flex gap-2">
                                <button
                                    onClick={handleCancel}
                                    className="px-4 py-2 border rounded-lg hover:bg-gray-100"
                                >
                                    ✕ Cancel
                                </button>
                                <button
                                    onClick={handleSave}
                                    disabled={isSaving}
                                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                                >
                                    💾 {isSaving ? 'Saving...' : 'Save'}
                                </button>
                            </div>
                        )}
                    </div>

                    <div className="space-y-6">
                        <div className="flex items-center gap-6">
                            <div className="w-24 h-24 rounded-full bg-gray-200 flex items-center justify-center overflow-hidden">
                                {(isEditing ? formData.avatar_url : user?.avatar_url) ? (
                                    <img
                                        src={isEditing ? formData.avatar_url : user?.avatar_url}
                                        alt="Avatar"
                                        className="w-full h-full object-cover"
                                    />
                                ) : (
                                    <span className="text-4xl font-bold text-gray-600">
                                        {user?.username?.charAt(0).toUpperCase()}
                                    </span>
                                )}
                            </div>
                            <div className="flex-1">
                                <h3 className="text-xl font-semibold">
                                    {user?.full_name || user?.username}
                                </h3>
                                <p className="text-gray-600">@{user?.username}</p>
                                <p className="text-sm text-gray-500">{user?.email}</p>
                            </div>
                        </div>

                        {isEditing ? (
                            <div className="space-y-4">
                                <div>
                                    <label htmlFor="full_name" className="block text-sm font-medium mb-2">
                                        Full Name
                                    </label>
                                    <input
                                        id="full_name"
                                        type="text"
                                        value={formData.full_name}
                                        onChange={(e) =>
                                            setFormData({ ...formData, full_name: e.target.value })
                                        }
                                        placeholder="Enter your full name"
                                        className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="bio" className="block text-sm font-medium mb-2">
                                        Bio
                                    </label>
                                    <textarea
                                        id="bio"
                                        value={formData.bio}
                                        onChange={(e) => setFormData({ ...formData, bio: e.target.value })}
                                        placeholder="Tell us about yourself"
                                        rows={4}
                                        className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="avatar_url" className="block text-sm font-medium mb-2">
                                        Avatar URL
                                    </label>
                                    <input
                                        id="avatar_url"
                                        type="url"
                                        value={formData.avatar_url}
                                        onChange={(e) =>
                                            setFormData({ ...formData, avatar_url: e.target.value })
                                        }
                                        placeholder="https://example.com/avatar.jpg"
                                        className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>
                            </div>
                        ) : (
                            <div className="space-y-4">
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500 mb-1">Bio</h4>
                                    <p>{user?.bio || 'No bio yet'}</p>
                                </div>

                                <div>
                                    <h4 className="text-sm font-medium text-gray-500 mb-1">
                                        Joined
                                    </h4>
                                    <p>{user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'Unknown'}</p>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </Layout>
    )
}
