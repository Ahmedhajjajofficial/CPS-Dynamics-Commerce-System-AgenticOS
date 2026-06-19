/**
 * Cart Component
 * ==============
 * Displays and manages the shopping cart.
 */

import React, { useState } from 'react';
import { Minus, Plus, Trash2, User, Tag, Receipt, CreditCard } from 'lucide-react';
import { useCartStore } from '../store/cartStore';
import { useSessionStore } from '../store/sessionStore';
import { formatCurrency, calculateItemTotal, calculateCartDiscountAmount } from '../utils/currency';

export const Cart: React.FC = () => {
  const { 
    cart, 
    removeItem, 
    updateQuantity, 
    updateDiscount, 
    setCartDiscount,
    clearCart 
  } = useCartStore();
  
  const { currentSession } = useSessionStore();
  const [showDiscountModal, setShowDiscountModal] = useState(false);
  const [itemDiscountId, setItemDiscountId] = useState<string | null>(null);

  const handleCheckout = () => {
    if (cart.items.length === 0) return;
    // TODO: Navigate to payment screen
    alert('Checkout - Total: ' + formatCurrency(cart.total));
  };

  const handleItemDiscount = (productId: string) => {
    setItemDiscountId(productId);
    setShowDiscountModal(true);
  };

  return (
    <div className="flex flex-col h-full bg-gray-50">
      {/* Header */}
      <div className="p-4 bg-white border-b">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-bold">Current Sale</h2>
          <div className="flex gap-2">
            {cart.items.length > 0 && (
              <button
                onClick={clearCart}
                className="p-2 text-red-500 hover:bg-red-50 rounded"
                title="Clear cart"
              >
                <Trash2 className="w-5 h-5" />
              </button>
            )}
          </div>
        </div>
        
        {cart.customer && (
          <div className="flex items-center gap-2 mt-2 p-2 bg-blue-50 rounded">
            <User className="w-4 h-4 text-blue-600" />
            <span className="text-sm text-blue-800">
              {cart.customer.name} ({cart.customer.membershipTier})
            </span>
            <span className="text-xs text-blue-600 ml-auto">
              {cart.customer.loyaltyPoints} points
            </span>
          </div>
        )}
      </div>

      {/* Cart Items */}
      <div className="flex-1 overflow-auto p-4">
        {cart.items.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-400">
            <Receipt className="w-16 h-16 mb-4" />
            <p>Cart is empty</p>
            <p className="text-sm">Scan or select products to add</p>
          </div>
        ) : (
          <div className="space-y-3">
            {cart.items.map(item => (
              <div key={item.product.id} className="bg-white p-3 rounded-lg shadow-sm">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h4 className="font-medium">{item.product.name}</h4>
                    <p className="text-sm text-gray-500">{formatCurrency(item.product.price)} each</p>
                    {item.notes && (
                      <p className="text-xs text-gray-400 mt-1">{item.notes}</p>
                    )}
                  </div>
                  <button
                    onClick={() => removeItem(item.product.id)}
                    className="p-1 text-red-400 hover:text-red-600"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>

                <div className="flex items-center justify-between mt-3">
                  {/* Quantity Controls */}
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => updateQuantity(item.product.id, item.quantity - 1)}
                      className="p-1 rounded bg-gray-100 hover:bg-gray-200"
                    >
                      <Minus className="w-4 h-4" />
                    </button>
                    <span className="w-8 text-center font-medium">{item.quantity}</span>
                    <button
                      onClick={() => updateQuantity(item.product.id, item.quantity + 1)}
                      className="p-1 rounded bg-gray-100 hover:bg-gray-200"
                    >
                      <Plus className="w-4 h-4" />
                    </button>
                  </div>

                  {/* Discount & Price */}
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => handleItemDiscount(item.product.id)}
                      className={`flex items-center gap-1 text-sm ${
                        item.discount > 0 ? 'text-green-600' : 'text-gray-400'
                      }`}
                    >
                      <Tag className="w-4 h-4" />
                      {item.discount > 0 && `${item.discount}%`}
                    </button>
                    <span className="font-bold">
                      {formatCurrency(calculateItemTotal(item))}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer - Totals */}
      {cart.items.length > 0 && (
        <div className="p-4 bg-white border-t">
          {/* Subtotal */}
          <div className="flex justify-between text-sm mb-1">
            <span className="text-gray-600">Subtotal</span>
            <span>{formatCurrency(cart.subtotal)}</span>
          </div>

          {/* Cart Discount */}
          {cart.discount > 0 && (
            <div className="flex justify-between text-sm mb-1 text-green-600">
              <span>Cart Discount ({cart.discount}%)</span>
              <span>-{formatCurrency(calculateCartDiscountAmount(cart.subtotal, cart.discount))}</span>
            </div>
          )}

          {/* Tax */}
          <div className="flex justify-between text-sm mb-2">
            <span className="text-gray-600">Tax</span>
            <span>{formatCurrency(cart.taxAmount)}</span>
          </div>

          {/* Total */}
          <div className="flex justify-between text-xl font-bold mb-4">
            <span>Total</span>
            <span className="text-blue-600">{formatCurrency(cart.total)}</span>
          </div>

          {/* Checkout Button */}
          <button
            onClick={handleCheckout}
            disabled={!currentSession}
            className="w-full py-3 bg-blue-600 text-white rounded-lg font-bold hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            <CreditCard className="w-5 h-5" />
            {currentSession ? 'Proceed to Payment' : 'Start Session First'}
          </button>
        </div>
      )}

      {/* Discount Modal */}
      {showDiscountModal && (
        <DiscountModal
          currentDiscount={itemDiscountId 
            ? cart.items.find(i => i.product.id === itemDiscountId)?.discount || 0
            : cart.discount
          }
          onApply={(discount) => {
            if (itemDiscountId) {
              updateDiscount(itemDiscountId, discount);
            } else {
              setCartDiscount(discount);
            }
            setShowDiscountModal(false);
            setItemDiscountId(null);
          }}
          onClose={() => {
            setShowDiscountModal(false);
            setItemDiscountId(null);
          }}
        />
      )}
    </div>
  );
};

// Discount Modal Component
interface DiscountModalProps {
  currentDiscount: number;
  onApply: (discount: number) => void;
  onClose: () => void;
}

const DiscountModal: React.FC<DiscountModalProps> = ({ 
  currentDiscount, 
  onApply, 
  onClose 
}) => {
  const [discount, setDiscount] = useState(currentDiscount);
  const presets = [5, 10, 15, 20, 25, 50];

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-80">
        <h3 className="text-lg font-bold mb-4">Apply Discount</h3>
        
        <div className="mb-4">
          <label className="block text-sm text-gray-600 mb-1">Discount %</label>
          <input
            type="number"
            min="0"
            max="100"
            value={discount}
            onChange={(e) => setDiscount(Number(e.target.value))}
            className="w-full p-2 border rounded"
          />
        </div>

        <div className="grid grid-cols-3 gap-2 mb-4">
          {presets.map(preset => (
            <button
              key={preset}
              onClick={() => setDiscount(preset)}
              className="p-2 bg-gray-100 rounded hover:bg-gray-200"
            >
              {preset}%
            </button>
          ))}
        </div>

        <div className="flex gap-2">
          <button
            onClick={onClose}
            className="flex-1 py-2 border rounded hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={() => onApply(discount)}
            className="flex-1 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Apply
          </button>
        </div>
      </div>
    </div>
  );
};
