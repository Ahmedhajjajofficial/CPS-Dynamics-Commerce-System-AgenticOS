/**
 * Session Store - Zustand
 * =======================
 * Manages the cashier session state with API integration.
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Session, User, SyncStatus } from '../types';
import { apiService } from '../services/api';

interface SessionState {
  // Current session
  currentSession: Session | null;
  currentUser: User | null;
  
  // Sync status
  syncStatus: SyncStatus;
  
  // Actions
  login: (user: User) => void;
  logout: () => void;
  startSession: (openingBalance: number, registerId: string) => Promise<void>;
  endSession: (closingBalance: number) => Promise<void>;
  updateSyncStatus: (status: Partial<SyncStatus>) => void;
  incrementTransactionCount: (amount: number) => void;
}

const initialSyncStatus: SyncStatus = {
  isOnline: true,
  pendingTransactions: 0
};

export const useSessionStore = create<SessionState>()(
  persist(
    (set, get) => ({
      currentSession: null,
      currentUser: null,
      syncStatus: initialSyncStatus,

      login: (user) => {
        set({ currentUser: user });
      },

      logout: () => {
        set({ 
          currentUser: null,
          currentSession: null 
        });
      },

      startSession: async (openingBalance, registerId) => {
        const { currentUser } = get();
        if (!currentUser) return;

        const response = await apiService.startSession({
          cashier_id: currentUser.id,
          register_id: registerId,
          opening_balance: openingBalance,
        });

        if (!response.success || !response.data) {
          throw new Error(response.error || 'Failed to start session');
        }

        const newSession: Session = {
          id: response.data.session_id,
          cashierId: response.data.cashier_id,
          cashierName: currentUser.name,
          registerId: response.data.register_id,
          branchId: 'BR001',
          openingBalance: response.data.opening_balance,
          startedAt: response.data.started_at,
          status: response.data.status as Session['status'],
          transactionCount: 0,
          totalSales: 0
        };

        set({ currentSession: newSession });
      },

      endSession: async (closingBalance) => {
        const { currentSession } = get();
        if (!currentSession) return;

        const response = await apiService.closeSession(currentSession.id, {
          closing_balance: closingBalance,
          total_sales: currentSession.totalSales,
          transaction_count: currentSession.transactionCount,
        });

        if (!response.success) {
          throw new Error(response.error || 'Failed to end session');
        }

        set({
          currentSession: {
            ...currentSession,
            closingBalance,
            endedAt: new Date().toISOString(),
            status: 'closed'
          }
        });
      },

      updateSyncStatus: (status) => {
        const { syncStatus } = get();
        set({
          syncStatus: { ...syncStatus, ...status }
        });
      },

      incrementTransactionCount: (amount) => {
        const { currentSession } = get();
        if (!currentSession) return;

        set({
          currentSession: {
            ...currentSession,
            transactionCount: currentSession.transactionCount + 1,
            totalSales: currentSession.totalSales + amount
          }
        });
      }
    }),
    {
      name: 'pos-session-storage',
      partialize: (state) => ({ 
        currentUser: state.currentUser,
        currentSession: state.currentSession 
      })
    }
  )
);
