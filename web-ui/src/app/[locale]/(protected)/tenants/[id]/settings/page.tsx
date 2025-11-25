'use client'
import { useEffect, useState, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import { useT } from '../../../../../../lib/i18n-context'
import { useNotification } from '../../../../../../contexts/NotificationContext'
import { getUser } from '../../../../../../lib/auth'
import AuthGuard from '../../../../../../components/AuthGuard'
import Footer from '../../../../../../components/Footer'
import {
    getTenantByID,
    getTenantMembers,
    updateTenant,
    type TenantWithMemberCount,
    type TenantMember,
    type TenantVisibility,
    type TenantAccessType,
    type UpdateTenantRequest,
} from '../../../../../../lib/api'

export default function TenantSettingsPage() {
    const router = useRouter()
    const params = useParams()
    const t = useT()
    const notification = useNotification()
    const tenantId = params.id as string
    const locale = params.locale as string

    const [tenant, setTenant] = useState<TenantWithMemberCount | null>(null)
    const [myMembership, setMyMembership] = useState<TenantMember | null>(null)
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)

    // Form state
    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [visibility, setVisibility] = useState<TenantVisibility>('private')
    const [accessType, setAccessType] = useState<TenantAccessType>('open')
    const [password, setPassword] = useState('')
    const [maxMembers, setMaxMembers] = useState<number>(100)

    const user = getUser()

    const loadTenantData = useCallback(async () => {
        setLoading(true)
        try {
            const tenantData = await getTenantByID(tenantId)
            setTenant(tenantData)

            // Populate form with current values
            setName(tenantData.name)
            setDescription(tenantData.description || '')
            setVisibility(tenantData.visibility)
            setAccessType(tenantData.access_type)
            setMaxMembers(tenantData.max_members || 100)

            const membersData = await getTenantMembers(tenantId)

            // Find my membership
            if (user) {
                const myMember = membersData.find((m: TenantMember) => m.user_id === user.id)
                setMyMembership(myMember || null)
            }
        } catch (err) {
            notification.error(t('error.generic'), String(err))
        } finally {
            setLoading(false)
        }
    }, [tenantId, user, notification, t])

    useEffect(() => {
        loadTenantData()
    }, [loadTenantData])

    // Permission check
    const canEditSettings = myMembership && ['owner', 'admin'].includes(myMembership.role)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!canEditSettings) {
            notification.error(t('error.generic'), 'You do not have permission to edit settings')
            return
        }

        setSaving(true)
        try {
            const updateReq: UpdateTenantRequest = {}

            // Only include changed fields
            if (name !== tenant?.name) updateReq.name = name
            if (description !== (tenant?.description || '')) updateReq.description = description
            if (visibility !== tenant?.visibility) updateReq.visibility = visibility
            if (accessType !== tenant?.access_type) updateReq.access_type = accessType
            if (maxMembers !== (tenant?.max_members || 100)) updateReq.max_members = maxMembers

            // Include password if changing to password access or updating password
            if (accessType === 'password' && password) {
                updateReq.password = password
            }

            // Validate password required for password access
            if (accessType === 'password' && !password && tenant?.access_type !== 'password') {
                notification.error(t('tenant.settings.title'), t('tenant.settings.passwordRequired'))
                setSaving(false)
                return
            }

            await updateTenant(tenantId, updateReq)
            notification.success(t('tenant.settings.title'), t('tenant.settings.saved'))
            setPassword('') // Clear password field after save
            await loadTenantData() // Refresh data
        } catch (err) {
            notification.error(t('tenant.settings.title'), String(err))
        } finally {
            setSaving(false)
        }
    }

    if (loading) {
        return (
            <AuthGuard>
                <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
                    <div className="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full" />
                </div>
            </AuthGuard>
        )
    }

    if (!tenant) {
        return (
            <AuthGuard>
                <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
                    <div className="text-center">
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
                            {t('tenant.notFound')}
                        </h1>
                        <Link
                            href={`/${locale}/tenants`}
                            className="text-blue-600 hover:text-blue-800 dark:text-blue-400"
                        >
                            {t('tenant.backToList')}
                        </Link>
                    </div>
                </div>
            </AuthGuard>
        )
    }

    if (!canEditSettings) {
        return (
            <AuthGuard>
                <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
                    <div className="text-center">
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
                            {t('tenant.settings.noPermission')}
                        </h1>
                        <Link
                            href={`/${locale}/tenants/${tenantId}`}
                            className="text-blue-600 hover:text-blue-800 dark:text-blue-400"
                        >
                            {t('tenant.backToDetail')}
                        </Link>
                    </div>
                </div>
            </AuthGuard>
        )
    }

    return (
        <AuthGuard>
            <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
                {/* Header */}
                <header className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
                    <div className="max-w-4xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
                        <div className="flex items-center gap-4">
                            <Link
                                href={`/${locale}/tenants/${tenantId}`}
                                className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                                aria-label={t('common.back')}
                            >
                                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                                </svg>
                            </Link>
                            <div>
                                <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
                                    {t('tenant.settings.title')}
                                </h1>
                                <p className="text-sm text-gray-500 dark:text-gray-400">
                                    {tenant.name}
                                </p>
                            </div>
                        </div>
                    </div>
                </header>

                {/* Main Content */}
                <main className="max-w-4xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
                    <form onSubmit={handleSubmit} className="space-y-8">
                        {/* Basic Information */}
                        <section className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
                            <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-6">
                                {t('tenant.settings.basicInfo')}
                            </h2>

                            <div className="space-y-6">
                                {/* Name */}
                                <div>
                                    <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        {t('tenant.name')} <span className="text-red-500">*</span>
                                    </label>
                                    <input
                                        type="text"
                                        id="name"
                                        value={name}
                                        onChange={(e) => setName(e.target.value)}
                                        required
                                        minLength={2}
                                        maxLength={100}
                                        className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    />
                                </div>

                                {/* Description */}
                                <div>
                                    <label htmlFor="description" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        {t('tenant.description')}
                                    </label>
                                    <textarea
                                        id="description"
                                        value={description}
                                        onChange={(e) => setDescription(e.target.value)}
                                        rows={4}
                                        maxLength={1000}
                                        className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                                        placeholder={t('tenant.settings.descriptionPlaceholder')}
                                    />
                                    <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                                        {description.length}/1000
                                    </p>
                                </div>

                                {/* Max Members */}
                                <div>
                                    <label htmlFor="maxMembers" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        {t('tenant.settings.maxMembers')}
                                    </label>
                                    <input
                                        type="number"
                                        id="maxMembers"
                                        value={maxMembers}
                                        onChange={(e) => setMaxMembers(parseInt(e.target.value) || 100)}
                                        min={1}
                                        max={10000}
                                        className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    />
                                    <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                                        {t('tenant.settings.maxMembersHelp')}
                                    </p>
                                </div>
                            </div>
                        </section>

                        {/* Privacy Settings */}
                        <section className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
                            <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-6">
                                {t('tenant.settings.privacy')}
                            </h2>

                            <div className="space-y-6">
                                {/* Visibility */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                                        {t('tenant.visibility')}
                                    </label>
                                    <div className="space-y-3">
                                        <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-600 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
                                            <input
                                                type="radio"
                                                name="visibility"
                                                value="public"
                                                checked={visibility === 'public'}
                                                onChange={() => setVisibility('public')}
                                                className="mt-1"
                                            />
                                            <div>
                                                <span className="block font-medium text-gray-900 dark:text-white">
                                                    {t('tenant.visibility.public')}
                                                </span>
                                                <span className="text-sm text-gray-500 dark:text-gray-400">
                                                    {t('tenant.visibility.publicDesc')}
                                                </span>
                                            </div>
                                        </label>
                                        <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-600 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
                                            <input
                                                type="radio"
                                                name="visibility"
                                                value="private"
                                                checked={visibility === 'private'}
                                                onChange={() => setVisibility('private')}
                                                className="mt-1"
                                            />
                                            <div>
                                                <span className="block font-medium text-gray-900 dark:text-white">
                                                    {t('tenant.visibility.private')}
                                                </span>
                                                <span className="text-sm text-gray-500 dark:text-gray-400">
                                                    {t('tenant.visibility.privateDesc')}
                                                </span>
                                            </div>
                                        </label>
                                    </div>
                                </div>

                                {/* Access Type */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                                        {t('tenant.accessType')}
                                    </label>
                                    <div className="space-y-3">
                                        <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-600 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
                                            <input
                                                type="radio"
                                                name="accessType"
                                                value="open"
                                                checked={accessType === 'open'}
                                                onChange={() => setAccessType('open')}
                                                className="mt-1"
                                            />
                                            <div>
                                                <span className="block font-medium text-gray-900 dark:text-white">
                                                    {t('tenant.accessType.open')}
                                                </span>
                                                <span className="text-sm text-gray-500 dark:text-gray-400">
                                                    {t('tenant.accessType.openDesc')}
                                                </span>
                                            </div>
                                        </label>
                                        <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-600 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
                                            <input
                                                type="radio"
                                                name="accessType"
                                                value="password"
                                                checked={accessType === 'password'}
                                                onChange={() => setAccessType('password')}
                                                className="mt-1"
                                            />
                                            <div>
                                                <span className="block font-medium text-gray-900 dark:text-white">
                                                    {t('tenant.accessType.password')}
                                                </span>
                                                <span className="text-sm text-gray-500 dark:text-gray-400">
                                                    {t('tenant.accessType.passwordDesc')}
                                                </span>
                                            </div>
                                        </label>
                                        <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-600 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50">
                                            <input
                                                type="radio"
                                                name="accessType"
                                                value="invite_only"
                                                checked={accessType === 'invite_only'}
                                                onChange={() => setAccessType('invite_only')}
                                                className="mt-1"
                                            />
                                            <div>
                                                <span className="block font-medium text-gray-900 dark:text-white">
                                                    {t('tenant.accessType.inviteOnly')}
                                                </span>
                                                <span className="text-sm text-gray-500 dark:text-gray-400">
                                                    {t('tenant.accessType.inviteOnlyDesc')}
                                                </span>
                                            </div>
                                        </label>
                                    </div>
                                </div>

                                {/* Password Field (shown when password access is selected) */}
                                {accessType === 'password' && (
                                    <div className="animate-in fade-in duration-200">
                                        <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                            {t('tenant.settings.tenantPassword')}
                                            {tenant.access_type !== 'password' && <span className="text-red-500"> *</span>}
                                        </label>
                                        <input
                                            type="password"
                                            id="password"
                                            value={password}
                                            onChange={(e) => setPassword(e.target.value)}
                                            required={tenant.access_type !== 'password'}
                                            className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                            placeholder={tenant.access_type === 'password' ? t('tenant.settings.leaveBlankToKeep') : t('tenant.settings.enterPassword')}
                                        />
                                        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                                            {tenant.access_type === 'password'
                                                ? t('tenant.settings.passwordChangeHint')
                                                : t('tenant.settings.passwordRequiredHint')
                                            }
                                        </p>
                                    </div>
                                )}
                            </div>
                        </section>

                        {/* Danger Zone */}
                        <section className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-red-200 dark:border-red-900/50 p-6">
                            <h2 className="text-lg font-medium text-red-600 dark:text-red-400 mb-4">
                                {t('tenant.settings.dangerZone')}
                            </h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                                {t('tenant.settings.dangerZoneDesc')}
                            </p>
                            <button
                                type="button"
                                className="px-4 py-2 text-sm font-medium text-red-600 dark:text-red-400 border border-red-300 dark:border-red-700 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                                onClick={() => {
                                    // TODO: Implement tenant deletion
                                    notification.error(t('tenant.settings.title'), 'Tenant deletion not yet implemented')
                                }}
                            >
                                {t('tenant.settings.deleteTenant')}
                            </button>
                        </section>

                        {/* Submit Button */}
                        <div className="flex justify-end gap-4">
                            <Link
                                href={`/${locale}/tenants/${tenantId}`}
                                className="px-6 py-2 text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                            >
                                {t('common.cancel')}
                            </Link>
                            <button
                                type="submit"
                                disabled={saving}
                                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                            >
                                {saving && (
                                    <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                                    </svg>
                                )}
                                {t('common.save')}
                            </button>
                        </div>
                    </form>
                </main>

                <Footer />
            </div>
        </AuthGuard>
    )
}
