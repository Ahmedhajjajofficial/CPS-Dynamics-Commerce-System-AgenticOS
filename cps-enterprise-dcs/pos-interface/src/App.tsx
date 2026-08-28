/**
 * RockDeals POS - Main Application
 * ==================================
 * The sovereign point of sale interface.
 */

import React, { useState, useEffect } from 'react';
import { ShoppingCart, Package, History, Settings, User, Wifi, WifiOff, LogOut } from 'lucide-react';
import { ProductGrid } from './components/ProductGrid';
import { Cart } from './components/Cart';
import { useSessionStore } from './store/sessionStore';
import { Product, User as UserType } from './types';
import { formatCurrency } from './utils/currency';
import { apiService } from './services/api';

type View = 'sale' | 'products' | 'history' | 'settings';

function App() {
  const [currentView, setCurrentView] = useState<View>('sale');
  const [showLogin, setShowLogin] = useState(true);
  const [showSessionStart, setShowSessionStart] = useState(false);
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [productsError, setProductsError] = useState<string | null>(null);
  
  const { 
    currentUser, 
    currentSession, 
    syncStatus, 
    login, 
    logout, 
    startSession,
    endSession,
    updateSyncStatus
  } = useSessionStore();

  // Load products from API
  useEffect(() => {
    const loadProducts = async () => {
      setLoadingProducts(true);
      setProductsError(null);
      const response = await apiService.getProducts();
      if (response.success && response.data) {
        setProducts(response.data);
        const uniqueCategories = Array.from(
          new Set(response.data.map((p: Product) => p.category).filter(Boolean))
        ) as string[];
        setCategories(uniqueCategories);
      } else {
        setProductsError(response.error || 'Failed to load products');
      }
      setLoadingProducts(false);
    };

    loadProducts();
  }, []);

  // Monitor connectivity
  useEffect(() => {
    const updateOnlineStatus = () => {
      updateSyncStatus({ isOnline: navigator.onLine });
    };

    updateOnlineStatus();
    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);

    return () => {
      window.removeEventListener('online', updateOnlineStatus);
      window.removeEventListener('offline', updateOnlineStatus);
    };
  }, [updateSyncStatus]);

  // Auto-login for demo
  useEffect(() => {
    if (!currentUser) {
      login({
        id: 'user1',
        name: 'John Cashier',
        email: 'john@rockdeals.com',
        role: 'cashier',
        permissions: ['sales', 'refunds'],
        isActive: true
      });
      setShowLogin(false);
      setShowSessionStart(true);
    }
  }, [currentUser, login]);

  const handleStartSession = async (openingBalance: number) => {
    try {
      await startSession(openingBalance, 'REG001');
      setShowSessionStart(false);
    } catch (error) {
      alert(error instanceof Error ? error.message : 'Failed to start session');
    }
  };

  const handleEndSession = async () => {
    if (currentSession) {
      try {
        await endSession(currentSession.totalSales);
        alert('Session ended successfully');
      } catch (error) {
        alert(error instanceof Error ? error.message : 'Failed to end session');
      }
    }
  };

  const handleLogout = () => {
    logout();
    setShowLogin(true);
  };

  // Login Screen
  if (showLogin) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-xl p-8 w-96">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-blue-600">RockDeals POS</h1>
            <p className="text-gray-500 mt-2">Sign in to continue</p>
          </div>
          
          <div className="space-y-4">
            <input
              type="text"
              placeholder="Employee ID"
              className="w-full p-3 border rounded-lg"
              defaultValue="EMP001"
            />
            <input
              type="password"
              placeholder="Password"
              className="w-full p-3 border rounded-lg"
            />
            <button
              onClick={() => {
                login({
                  id: 'user1',
                  name: 'John Cashier',
                  email: 'john@rockdeals.com',
                  role: 'cashier',
                  permissions: ['sales', 'refunds'],
                  isActive: true
                });
                setShowLogin(false);
                setShowSessionStart(true);
              }}
              className="w-full py-3 bg-blue-600 text-white rounded-lg font-bold hover:bg-blue-700"
            >
              Sign In
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Session Start Screen
  if (showSessionStart) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-xl p-8 w-96">
          <h2 className="text-2xl font-bold mb-4">Start Session</h2>
          <p className="text-gray-500 mb-6">Enter opening cash balance</p>
          
          <form onSubmit={(e) => {
            e.preventDefault();
            const formData = new FormData(e.currentTarget);
            const balance = Number(formData.get('balance'));
            handleStartSession(balance);
          }}>
            <input
              name="balance"
              type="number"
              step="0.01"
              placeholder="0.00"
              className="w-full p-3 border rounded-lg mb-4 text-2xl text-center"
              autoFocus
            />
            <button
              type="submit"
              className="w-full py-3 bg-blue-600 text-white rounded-lg font-bold hover:bg-blue-700"
            >
              Start Session
            </button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col bg-gray-100">
      {/* Header */}
      <header className="bg-white border-b px-4 py-2 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold text-blue-600">RockDeals POS</h1>
          
          {/* Session Info */}
          {currentSession && (
            <div className="flex items-center gap-4 text-sm">
              <span className="px-2 py-1 bg-green-100 text-green-700 rounded">
                Session Active
              </span>
              <span className="text-gray-500">
                Register: {currentSession.registerId}
              </span>
              <span className="text-gray-500">
                Sales: {formatCurrency(currentSession.totalSales)}
              </span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-4">
          {/* Sync Status */}
          <div className={`flex items-center gap-1 text-sm ${syncStatus.isOnline ? 'text-green-600' : 'text-red-500'}`}>
            {syncStatus.isOnline ? <Wifi className="w-4 h-4" /> : <WifiOff className="w-4 h-4" />}
            {syncStatus.isOnline ? 'Online' : 'Offline'}
          </div>

          {/* User */}
          <div className="flex items-center gap-2">
            <User className="w-5 h-5 text-gray-500" />
            <span className="text-sm">{currentUser?.name}</span>
          </div>

          {/* Logout */}
          <button
            onClick={handleLogout}
            className="p-2 text-gray-500 hover:text-red-500"
            title="Logout"
          >
            <LogOut className="w-5 h-5" />
          </button>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <nav className="w-16 bg-white border-r flex flex-col items-center py-4 gap-2">
          <NavButton 
            icon={<ShoppingCart className="w-5 h-5" />} 
            label="Sale"
            active={currentView === 'sale'}
            onClick={() => setCurrentView('sale')}
          />
          <NavButton 
            icon={<Package className="w-5 h-5" />} 
            label="Products"
            active={currentView === 'products'}
            onClick={() => setCurrentView('products')}
          />
          <NavButton 
            icon={<History className="w-5 h-5" />} 
            label="History"
            active={currentView === 'history'}
            onClick={() => setCurrentView('history')}
          />
          <NavButton 
            icon={<Settings className="w-5 h-5" />} 
            label="Settings"
            active={currentView === 'settings'}
            onClick={() => setCurrentView('settings')}
          />
        </nav>

        {/* Content Area */}
        <main className="flex-1 flex overflow-hidden">
          {currentView === 'sale' && (
            <>
              {/* Product Grid */}
              <div className="flex-1">
                {loadingProducts ? (
                  <div className="flex items-center justify-center h-full text-gray-500">
                    Loading products...
                  </div>
                ) : productsError ? (
                  <div className="flex items-center justify-center h-full text-red-500">
                    {productsError}
                  </div>
                ) : (
                  <ProductGrid products={products} categories={categories} />
                )}
              </div>
               
              {/* Cart */}
              <div className="w-96 border-l">
                <Cart />
              </div>
            </>
          )}

          {currentView === 'products' && (
            <div className="flex-1 p-8">
              <h2 className="text-2xl font-bold mb-4">Product Management</h2>
              <p className="text-gray-500">Product management interface coming soon...</p>
            </div>
          )}

          {currentView === 'history' && (
            <div className="flex-1 p-8">
              <h2 className="text-2xl font-bold mb-4">Transaction History</h2>
              <p className="text-gray-500">Transaction history coming soon...</p>
            </div>
          )}

          {currentView === 'settings' && (
            <div className="flex-1 p-8">
              <h2 className="text-2xl font-bold mb-4">Settings</h2>
              
              <div className="space-y-4 max-w-md">
                <div className="bg-white p-4 rounded-lg shadow">
                  <h3 className="font-bold mb-2">Session</h3>
                  {currentSession && (
                    <button
                      onClick={handleEndSession}
                      className="w-full py-2 bg-red-500 text-white rounded hover:bg-red-600"
                    >
                      End Session
                    </button>
                  )}
                </div>

                <div className="bg-white p-4 rounded-lg shadow">
                  <h3 className="font-bold mb-2">Sync Status</h3>
                  <div className="text-sm text-gray-600">
                    <p>Pending transactions: {syncStatus.pendingTransactions}</p>
                    {syncStatus.lastSyncAt && (
                      <p>Last sync: {new Date(syncStatus.lastSyncAt).toLocaleString()}</p>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

// Navigation Button Component
interface NavButtonProps {
  icon: React.ReactNode;
  label: string;
  active?: boolean;
  onClick: () => void;
}

const NavButton: React.FC<NavButtonProps> = ({ icon, label, active, onClick }) => (
  <button
    onClick={onClick}
    className={`p-3 rounded-lg flex flex-col items-center gap-1 w-14 ${
      active 
        ? 'bg-blue-100 text-blue-600' 
        : 'text-gray-500 hover:bg-gray-100'
    }`}
    title={label}
  >
    {icon}
    <span className="text-xs">{label}</span>
  </button>
);

export default App;
