/**
 * CPS Enterprise DCS - Main Application
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import { useState } from 'react'
import {
  GlassNavbar,
  GlassSidebar,
  GlassMetricCard,
  GlassCard,
  GlassCardHeader,
  GlassCardTitle,
  GlassCardDescription,
  GlassCardContent,
  GlassCardFooter,
  GlassButton,
  GlassBadge,
  GlassTable,
  GlassProgress,
} from '@/components/glass'
import type { SidebarGroup, Column } from '@/components/glass'
import './App.css'

/* ============================================================
   Sidebar Navigation Configuration
   ============================================================ */
const sidebarGroups: SidebarGroup[] = [
  {
    title: "الرئيسية",
    items: [
      {
        id: "dashboard",
        label: "لوحة التحكم",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><rect x="3" y="3" width="7" height="7" rx="1" /><rect x="14" y="3" width="7" height="7" rx="1" /><rect x="3" y="14" width="7" height="7" rx="1" /><rect x="14" y="14" width="7" height="7" rx="1" /></svg>,
      },
      {
        id: "analytics",
        label: "التحليلات",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="18" y1="20" x2="18" y2="10" /><line x1="12" y1="20" x2="12" y2="4" /><line x1="6" y1="20" x2="6" y2="14" /></svg>,
        badge: "جديد",
      },
    ],
  },
  {
    title: "العمليات",
    items: [
      {
        id: "pos",
        label: "نقاط البيع",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><rect x="2" y="3" width="20" height="14" rx="2" /><line x1="8" y1="21" x2="16" y2="21" /><line x1="12" y1="17" x2="12" y2="21" /></svg>,
        badge: 5,
      },
      {
        id: "inventory",
        label: "المخزون",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" /><polyline points="3.27 6.96 12 12.01 20.73 6.96" /><line x1="12" y1="22.08" x2="12" y2="12" /></svg>,
      },
      {
        id: "orders",
        label: "الطلبات",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /><line x1="16" y1="13" x2="8" y2="13" /><line x1="16" y1="17" x2="8" y2="17" /></svg>,
      },
    ],
  },
  {
    title: "النظام",
    items: [
      {
        id: "agents",
        label: "الوكلاء",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" /></svg>,
      },
      {
        id: "sync",
        label: "المزامنة",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><polyline points="23 4 23 10 17 10" /><polyline points="1 20 1 14 7 14" /><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" /></svg>,
      },
      {
        id: "settings",
        label: "الإعدادات",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" /><circle cx="12" cy="12" r="3" /></svg>,
      },
    ],
  },
]

/* ============================================================
   Sample Data
   ============================================================ */
interface Transaction {
  id: string
  branch: string
  amount: string
  status: string
  time: string
  [key: string]: unknown
}

const recentTransactions: Transaction[] = [
  { id: "#TXN-4821", branch: "فرع الرياض الرئيسي", amount: "12,450 ر.س", status: "مكتمل", time: "منذ 3 دقائق" },
  { id: "#TXN-4820", branch: "فرع جدة", amount: "8,320 ر.س", status: "معالجة", time: "منذ 7 دقائق" },
  { id: "#TXN-4819", branch: "فرع الدمام", amount: "5,100 ر.س", status: "مكتمل", time: "منذ 12 دقيقة" },
  { id: "#TXN-4818", branch: "فرع المدينة", amount: "3,750 ر.س", status: "مكتمل", time: "منذ 18 دقيقة" },
  { id: "#TXN-4817", branch: "فرع الرياض الرئيسي", amount: "22,000 ر.س", status: "مراجعة", time: "منذ 25 دقيقة" },
]

const transactionColumns: Column<Transaction>[] = [
  { key: "id", header: "رقم العملية", render: (item) => <span className="text-mono text-[#00e5ff] font-medium">{item.id}</span> },
  { key: "branch", header: "الفرع", render: (item) => <span className="text-slate-300">{item.branch}</span> },
  { key: "amount", header: "المبلغ", render: (item) => <span className="text-white font-semibold text-mono">{item.amount}</span> },
  {
    key: "status",
    header: "الحالة",
    render: (item) => {
      const v = item.status === "مكتمل" ? "success" : item.status === "معالجة" ? "warning" : "info"
      return <GlassBadge variant={v} dot>{item.status}</GlassBadge>
    },
  },
  { key: "time", header: "الوقت", render: (item) => <span className="text-slate-500 text-xs">{item.time}</span> },
]

/* ============================================================
   App Component
   ============================================================ */
function App() {
  const [activeNav, setActiveNav] = useState("dashboard")
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  return (
    <div className="dark min-h-screen bg-[#020617] text-slate-100" dir="rtl">
      {/* Animated Background */}
      <div className="fixed inset-0 z-0 pointer-events-none">
        <div className="bg-mesh w-full h-full" />
        <div className="bg-grid w-full h-full absolute inset-0 opacity-40" />
        {/* Floating Orbs */}
        <div className="absolute top-[15%] right-[10%] w-[500px] h-[500px] rounded-full bg-[rgba(0,229,255,0.03)] blur-[100px] animate-float" />
        <div className="absolute bottom-[20%] left-[15%] w-[400px] h-[400px] rounded-full bg-[rgba(124,58,237,0.03)] blur-[100px] animate-float" style={{ animationDelay: '3s' }} />
      </div>

      {/* Layout */}
      <div className="relative z-10 flex flex-col min-h-screen">
        {/* Navbar */}
        <GlassNavbar onMenuToggle={() => setSidebarCollapsed(!sidebarCollapsed)} />

        <div className="flex flex-1">
          {/* Sidebar */}
          <GlassSidebar
            groups={sidebarGroups}
            collapsed={sidebarCollapsed}
            activeId={activeNav}
            onItemClick={setActiveNav}
          />

          {/* Main Content */}
          <main className="flex-1 p-6 overflow-y-auto">
            {/* Page Header */}
            <div className="mb-8 animate-fade-up opacity-0" style={{ animationFillMode: 'forwards' }}>
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-2xl font-bold text-gradient-light text-display">لوحة التحكم</h2>
                  <p className="text-sm text-slate-500 mt-1 text-brand">مرحباً بك في نظام التجارة الديناميكي المؤسسي</p>
                </div>
                <div className="flex items-center gap-3">
                  <GlassButton variant="outline" size="sm">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="7 10 12 15 17 10" /><line x1="12" y1="15" x2="12" y2="3" /></svg>
                    تصدير
                  </GlassButton>
                  <GlassButton variant="glow" size="sm">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" /></svg>
                    عملية جديدة
                  </GlassButton>
                </div>
              </div>
            </div>

            {/* Metrics Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
              <GlassMetricCard
                title="إجمالي المبيعات"
                value="284,520 ر.س"
                change="+12.5%"
                changeType="positive"
                subtitle="مقارنة بالأمس"
                animationDelay="0.1s"
                icon={<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="12" y1="1" x2="12" y2="23" /><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" /></svg>}
              />
              <GlassMetricCard
                title="العمليات النشطة"
                value="1,847"
                change="+8.2%"
                changeType="positive"
                subtitle="هذا الأسبوع"
                animationDelay="0.2s"
                iconColor="text-emerald-400"
                icon={<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12" /></svg>}
              />
              <GlassMetricCard
                title="الفروع المتصلة"
                value="24 / 26"
                change="92%"
                changeType="positive"
                subtitle="نسبة الاتصال"
                animationDelay="0.3s"
                iconColor="text-indigo-400"
                icon={<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M5.5 8.5 9 12l-3.5 3.5L2 12l3.5-3.5Z" /><path d="m12 2 3.5 3.5L12 9 8.5 5.5 12 2Z" /><path d="M18.5 8.5 22 12l-3.5 3.5L15 12l3.5-3.5Z" /><path d="m12 15 3.5 3.5L12 22l-3.5-3.5L12 15Z" /></svg>}
              />
              <GlassMetricCard
                title="وقت الاستجابة"
                value="45ms"
                change="-15ms"
                changeType="positive"
                subtitle="تحسن ملحوظ"
                animationDelay="0.4s"
                iconColor="text-amber-400"
                icon={<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" /></svg>}
              />
            </div>

            {/* Content Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-5 mb-8">
              {/* Transactions Table */}
              <GlassCard variant="default" noPadding className="lg:col-span-2 animate-fade-up opacity-0" style={{ animationDelay: '0.3s', animationFillMode: 'forwards' }}>
                <GlassCardHeader className="px-6 pt-6">
                  <GlassCardTitle>آخر العمليات</GlassCardTitle>
                  <GlassCardDescription>العمليات التجارية في الوقت الفعلي عبر جميع الفروع</GlassCardDescription>
                </GlassCardHeader>
                <GlassCardContent>
                  <GlassTable columns={transactionColumns} data={recentTransactions} />
                </GlassCardContent>
                <GlassCardFooter className="px-6 pb-5">
                  <GlassButton variant="ghost" size="sm">عرض جميع العمليات</GlassButton>
                </GlassCardFooter>
              </GlassCard>

              {/* System Status */}
              <GlassCard variant="default" className="animate-fade-up opacity-0" style={{ animationDelay: '0.4s', animationFillMode: 'forwards' }}>
                <GlassCardHeader>
                  <GlassCardTitle>حالة النظام</GlassCardTitle>
                  <GlassCardDescription>أداء الوكلاء والخدمات</GlassCardDescription>
                </GlassCardHeader>
                <GlassCardContent className="space-y-5">
                  <GlassProgress value={92} label="الوكيل المحلي" showValue color="cyan" />
                  <GlassProgress value={87} label="الوكيل الإقليمي" showValue color="emerald" />
                  <GlassProgress value={95} label="مزامنة CRDT" showValue color="indigo" />
                  <GlassProgress value={78} label="قاعدة البيانات" showValue color="amber" />
                  <GlassProgress value={99} label="التشفير (AES-256)" showValue color="cyan" />

                  <div className="pt-3 border-t border-white/[0.04] space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-slate-400 text-brand">وضع عدم الاتصال</span>
                      <GlassBadge variant="success" dot>جاهز</GlassBadge>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-slate-400 text-brand">آخر مزامنة</span>
                      <span className="text-xs text-slate-300 text-mono">منذ 30 ثانية</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-slate-400 text-brand">الأحداث المعلقة</span>
                      <span className="text-xs text-[#00e5ff] font-semibold text-mono">0</span>
                    </div>
                  </div>
                </GlassCardContent>
              </GlassCard>
            </div>

            {/* Bottom Row */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
              {/* Quick Actions */}
              <GlassCard variant="subtle" interactive className="animate-fade-up opacity-0" style={{ animationDelay: '0.5s', animationFillMode: 'forwards' }}>
                <GlassCardHeader>
                  <GlassCardTitle>إجراءات سريعة</GlassCardTitle>
                </GlassCardHeader>
                <GlassCardContent className="space-y-2">
                  <GlassButton variant="secondary" className="w-full justify-start gap-3">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" /></svg>
                    إضافة منتج جديد
                  </GlassButton>
                  <GlassButton variant="secondary" className="w-full justify-start gap-3">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" /></svg>
                    إضافة مستخدم
                  </GlassButton>
                  <GlassButton variant="secondary" className="w-full justify-start gap-3">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><polyline points="23 4 23 10 17 10" /><polyline points="1 20 1 14 7 14" /><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" /></svg>
                    مزامنة يدوية
                  </GlassButton>
                  <GlassButton variant="secondary" className="w-full justify-start gap-3">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /></svg>
                    تقرير يومي
                  </GlassButton>
                </GlassCardContent>
              </GlassCard>

              {/* Branch Performance */}
              <GlassCard variant="subtle" interactive className="animate-fade-up opacity-0" style={{ animationDelay: '0.55s', animationFillMode: 'forwards' }}>
                <GlassCardHeader>
                  <GlassCardTitle>أداء الفروع</GlassCardTitle>
                </GlassCardHeader>
                <GlassCardContent className="space-y-4">
                  {[
                    { name: "فرع الرياض الرئيسي", value: 94, color: "cyan" as const },
                    { name: "فرع جدة", value: 87, color: "emerald" as const },
                    { name: "فرع الدمام", value: 91, color: "indigo" as const },
                    { name: "فرع المدينة", value: 76, color: "amber" as const },
                  ].map((branch) => (
                    <GlassProgress
                      key={branch.name}
                      value={branch.value}
                      label={branch.name}
                      showValue
                      color={branch.color}
                      size="sm"
                    />
                  ))}
                </GlassCardContent>
              </GlassCard>

              {/* Security Status */}
              <GlassCard variant="subtle" interactive className="animate-fade-up opacity-0" style={{ animationDelay: '0.6s', animationFillMode: 'forwards' }}>
                <GlassCardHeader>
                  <GlassCardTitle>الأمان والتشفير</GlassCardTitle>
                </GlassCardHeader>
                <GlassCardContent className="space-y-3">
                  <div className="flex items-center gap-3 p-3 rounded-xl bg-white/[0.02] border border-white/[0.04]">
                    <div className="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className="text-emerald-400"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" /></svg>
                    </div>
                    <div className="flex-1">
                      <p className="text-xs font-medium text-slate-300 text-brand">AES-256-GCM</p>
                      <p className="text-[10px] text-slate-500">تشفير نشط</p>
                    </div>
                    <GlassBadge variant="success" size="sm">فعال</GlassBadge>
                  </div>
                  <div className="flex items-center gap-3 p-3 rounded-xl bg-white/[0.02] border border-white/[0.04]">
                    <div className="w-8 h-8 rounded-lg bg-indigo-500/10 flex items-center justify-center">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className="text-indigo-400"><rect x="3" y="11" width="18" height="11" rx="2" ry="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></svg>
                    </div>
                    <div className="flex-1">
                      <p className="text-xs font-medium text-slate-300 text-brand">HashiCorp Vault</p>
                      <p className="text-[10px] text-slate-500">إدارة المفاتيح</p>
                    </div>
                    <GlassBadge variant="success" size="sm">متصل</GlassBadge>
                  </div>
                  <div className="flex items-center gap-3 p-3 rounded-xl bg-white/[0.02] border border-white/[0.04]">
                    <div className="w-8 h-8 rounded-lg bg-cyan-500/10 flex items-center justify-center">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className="text-cyan-400"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" /></svg>
                    </div>
                    <div className="flex-1">
                      <p className="text-xs font-medium text-slate-300 text-brand">mTLS</p>
                      <p className="text-[10px] text-slate-500">اتصال آمن بين الوكلاء</p>
                    </div>
                    <GlassBadge variant="success" size="sm">نشط</GlassBadge>
                  </div>
                </GlassCardContent>
              </GlassCard>
            </div>

            {/* Footer */}
            <footer className="mt-12 pt-6 border-t border-white/[0.04] text-center">
              <p className="text-[11px] text-slate-600 text-brand">
                &copy; 2026 CP'S Enterprise DCS. جميع الحقوق محفوظة.
              </p>
              <p className="text-[10px] text-slate-700 mt-1">
                Developed by Ahmed Hajjaj &mdash; CP'S Enterprise Tech Solutions &mdash; CP'S Businesses
              </p>
            </footer>
          </main>
        </div>
      </div>
    </div>
  )
}

export default App
