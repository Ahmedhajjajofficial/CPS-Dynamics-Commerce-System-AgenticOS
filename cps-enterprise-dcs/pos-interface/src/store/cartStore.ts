/**
 * Cart Store - Zustand
 * ====================
 * Manages the shopping cart state.
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Cart, CartItem, Product, Customer } from '../types';
import { calculateItemTotal, calculateCartDiscountAmount } from '../utils/currency';

interface CartState {
  cart: Cart;
  
  // Actions
  addItem: (product: Product, quantity?: number) => void;
  removeItem: (productId: string) => void;
  updateQuantity: (productId: string, quantity: number) => void;
  updateDiscount: (productId: string, discount: number) => void;
  setCustomer: (customer: Customer | undefined) => void;
  setCartDiscount: (discount: number) => void;
  clearCart: () => void;
  addNote: (productId: string, note: string) => void;
}

const calculateCartTotals = (items: CartItem[], cartDiscount: number): Pick<Cart, 'subtotal' | 'taxAmount' | 'total'> => {
  const subtotal = items.reduce((sum, item) => sum + calculateItemTotal(item), 0);
  const afterCartDiscount = subtotal - calculateCartDiscountAmount(subtotal, cartDiscount);

  const taxAmount = items.reduce((sum, item) => {
    const taxableAmount = calculateItemTotal(item);
    return sum + (taxableAmount * (item.product.taxRate / 100));
  }, 0);

  return {
    subtotal,
    taxAmount,
    total: afterCartDiscount + taxAmount
  };
};

const initialCart: Cart = {
  items: [],
  discount: 0,
  taxAmount: 0,
  subtotal: 0,
  total: 0
};

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      cart: initialCart,

      addItem: (product, quantity = 1) => {
        const { cart } = get();
        const existingItem = cart.items.find(item => item.product.id === product.id);

        let newItems: CartItem[];
        if (existingItem) {
          newItems = cart.items.map(item =>
            item.product.id === product.id
              ? { ...item, quantity: item.quantity + quantity }
              : item
          );
        } else {
          newItems = [...cart.items, { product, quantity, discount: 0 }];
        }

        const totals = calculateCartTotals(newItems, cart.discount);
        set({ cart: { ...cart, items: newItems, ...totals } });
      },

      removeItem: (productId) => {
        const { cart } = get();
        const newItems = cart.items.filter(item => item.product.id !== productId);
        const totals = calculateCartTotals(newItems, cart.discount);
        set({ cart: { ...cart, items: newItems, ...totals } });
      },

      updateQuantity: (productId, quantity) => {
        const { cart } = get();
        if (quantity <= 0) {
          get().removeItem(productId);
          return;
        }

        const newItems = cart.items.map(item =>
          item.product.id === productId ? { ...item, quantity } : item
        );
        const totals = calculateCartTotals(newItems, cart.discount);
        set({ cart: { ...cart, items: newItems, ...totals } });
      },

      updateDiscount: (productId, discount) => {
        const { cart } = get();
        const newItems = cart.items.map(item =>
          item.product.id === productId ? { ...item, discount } : item
        );
        const totals = calculateCartTotals(newItems, cart.discount);
        set({ cart: { ...cart, items: newItems, ...totals } });
      },

      setCustomer: (customer) => {
        const { cart } = get();
        set({ cart: { ...cart, customer } });
      },

      setCartDiscount: (discount) => {
        const { cart } = get();
        const totals = calculateCartTotals(cart.items, discount);
        set({ cart: { ...cart, discount, ...totals } });
      },

      clearCart: () => {
        set({ cart: initialCart });
      },

      addNote: (productId, note) => {
        const { cart } = get();
        const newItems = cart.items.map(item =>
          item.product.id === productId ? { ...item, notes: note } : item
        );
        set({ cart: { ...cart, items: newItems } });
      }
    }),
    {
      name: 'pos-cart-storage',
      partialize: (state) => ({ cart: state.cart })
    }
  )
);
