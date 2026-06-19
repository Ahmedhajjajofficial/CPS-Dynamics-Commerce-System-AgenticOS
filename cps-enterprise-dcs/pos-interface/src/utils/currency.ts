/**
 * Currency & Pricing Utilities
 * ============================
 * Shared helpers for formatting currency values and computing line-item
 * totals with discounts.  Eliminates the duplicated `$${x.toFixed(2)}`
 * and `price * qty * (1 - discount/100)` patterns scattered across
 * Cart, ProductGrid, and App components.
 */

import type { CartItem } from '../types';

/**
 * Format a number as a USD currency string.
 *
 * @example formatCurrency(29.9)  // "$29.90"
 * @example formatCurrency(1000)  // "$1,000.00"
 */
export function formatCurrency(amount: number): string {
  return `$${amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

/**
 * Compute the net total for a single cart line item after its per-item
 * discount is applied:  `price * quantity * (1 - discount / 100)`
 */
export function calculateItemTotal(item: CartItem): number {
  return item.product.price * item.quantity * (1 - item.discount / 100);
}

/**
 * Compute the cart-level discount amount from the subtotal.
 */
export function calculateCartDiscountAmount(subtotal: number, discountPercent: number): number {
  return subtotal * (discountPercent / 100);
}
