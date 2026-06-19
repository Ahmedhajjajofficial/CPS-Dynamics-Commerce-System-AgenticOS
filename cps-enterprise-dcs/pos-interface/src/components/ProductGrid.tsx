/**
 * Product Grid Component
 * ======================
 * Displays products in a grid layout for quick selection.
 */

import React, { useState, useMemo } from 'react';
import { Search, Grid3X3, List } from 'lucide-react';
import { Product } from '../types';
import { useCartStore } from '../store/cartStore';
import { formatCurrency } from '../utils/currency';

interface ProductGridProps {
  products: Product[];
  categories?: string[];
}

export const ProductGrid: React.FC<ProductGridProps> = ({ 
  products, 
  categories = [] 
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  
  const addItem = useCartStore(state => state.addItem);

  const filteredProducts = useMemo(() => {
    return products.filter(product => {
      const matchesSearch = 
        product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        product.sku.toLowerCase().includes(searchQuery.toLowerCase()) ||
        product.barcode?.includes(searchQuery);
      
      const matchesCategory = 
        selectedCategory === 'all' || product.category === selectedCategory;
      
      return matchesSearch && matchesCategory && product.isActive;
    });
  }, [products, searchQuery, selectedCategory]);

  const handleProductClick = (product: Product) => {
    addItem(product, 1);
  };

  return (
    <div className="flex flex-col h-full bg-white">
      {/* Header */}
      <div className="p-4 border-b space-y-3">
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search products..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        {/* Categories & View Mode */}
        <div className="flex items-center justify-between">
          <div className="flex gap-2 overflow-x-auto">
            <button
              onClick={() => setSelectedCategory('all')}
              className={`px-3 py-1 rounded-full text-sm whitespace-nowrap ${
                selectedCategory === 'all'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              All
            </button>
            {categories.map(category => (
              <button
                key={category}
                onClick={() => setSelectedCategory(category)}
                className={`px-3 py-1 rounded-full text-sm whitespace-nowrap ${
                  selectedCategory === category
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {category}
              </button>
            ))}
          </div>

          <div className="flex gap-1">
            <button
              onClick={() => setViewMode('grid')}
              className={`p-2 rounded ${viewMode === 'grid' ? 'bg-blue-100 text-blue-600' : 'text-gray-500'}`}
            >
              <Grid3X3 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`p-2 rounded ${viewMode === 'list' ? 'bg-blue-100 text-blue-600' : 'text-gray-500'}`}
            >
              <List className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Products */}
      <div className="flex-1 overflow-auto p-4">
        {filteredProducts.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            No products found
          </div>
        ) : viewMode === 'grid' ? (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {filteredProducts.map(product => (
              <button
                key={product.id}
                onClick={() => handleProductClick(product)}
                className="flex flex-col p-4 border rounded-lg hover:shadow-md hover:border-blue-300 transition-all text-left"
              >
                {product.imageUrl && (
                  <img
                    src={product.imageUrl}
                    alt={product.name}
                    className="w-full h-24 object-cover rounded mb-2"
                  />
                )}
                <h3 className="font-medium text-sm line-clamp-2">{product.name}</h3>
                <p className="text-gray-500 text-xs">{product.sku}</p>
                <p className="text-blue-600 font-bold mt-auto">{formatCurrency(product.price)}</p>
                {product.stockQuantity < 10 && (
                  <span className="text-xs text-red-500 mt-1">
                    Only {product.stockQuantity} left
                  </span>
                )}
              </button>
            ))}
          </div>
        ) : (
          <div className="space-y-2">
            {filteredProducts.map(product => (
              <button
                key={product.id}
                onClick={() => handleProductClick(product)}
                className="flex items-center w-full p-3 border rounded-lg hover:shadow-md hover:border-blue-300 transition-all text-left"
              >
                {product.imageUrl && (
                  <img
                    src={product.imageUrl}
                    alt={product.name}
                    className="w-12 h-12 object-cover rounded mr-3"
                  />
                )}
                <div className="flex-1">
                  <h3 className="font-medium">{product.name}</h3>
                  <p className="text-gray-500 text-sm">{product.sku} • {product.category}</p>
                </div>
                <div className="text-right">
                  <p className="text-blue-600 font-bold">{formatCurrency(product.price)}</p>
                  <p className="text-xs text-gray-500">Stock: {product.stockQuantity}</p>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
