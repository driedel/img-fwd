/**
 * img-fwd Demo — Main JavaScript
 * Features:
 *  - Measure and display image file sizes
 *  - Calculate savings percentages
 *  - Show real-time loading stats
 */

(function () {
  'use strict';

  const PROXY_BASE = 'https://img-fwd.driedel.dev';

  // ========================================
  // Utilities
  // ========================================

  function formatBytes(bytes) {
    if (bytes === 0 || bytes == null || isNaN(bytes)) return '—';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function calculateSavings(original, optimized) {
    if (!original || !optimized || original === 0) return null;
    const savings = Math.round(((original - optimized) / original) * 100);
    return savings;
  }

  // ========================================
  // Measure image sizes via Performance API
  // ========================================

  async function measureImageSize(url) {
    return new Promise((resolve) => {
      // Try Performance API first
      const entries = performance.getEntriesByType('resource');
      const found = entries.find((e) => e.name === url);
      if (found && found.transferSize > 0) {
        resolve(found.transferSize);
        return;
      }

      // Fallback: HEAD request with cache-busting
      const cacheBuster = url.includes('?') ? '&' : '?';
      const fetchUrl = url + cacheBuster + '_t=' + Date.now();
      
      fetch(fetchUrl, { method: 'HEAD', mode: 'cors' })
        .then((res) => {
          const length = res.headers.get('content-length');
          resolve(length ? parseInt(length, 10) : null);
        })
        .catch(() => resolve(null));
    });
  }

  // ========================================
  // Global function for inline onload
  // ========================================

  window.measureSize = function(img) {
    const container = img.closest('.resize-item, .effects-item, .format-item, .gallery-item');
    if (!container) return;
    
    const sizeEl = container.querySelector('.resize-size, .gallery-size');
    if (!sizeEl) return;
    
    const url = img.src;
    measureImageSize(url).then(size => {
      if (size) {
        sizeEl.textContent = formatBytes(size);
        sizeEl.style.color = 'var(--success)';
      }
    });
  };

  // ========================================
  // File Comparison Section
  // ========================================

  async function updateFileComparisons() {
    const comparisons = [
      { 
        name: 'photo', 
        origSize: 2346543, // bytes (2.3 MB)
        avifUrl: PROXY_BASE + '/images/photo.jpg?rs=1200',
        webpUrl: PROXY_BASE + '/images/photo.jpg?f=webp&rs=1200'
      },
      { 
        name: 'portrait', 
        origSize: 4178018, // bytes (4.1 MB)
        avifUrl: PROXY_BASE + '/images/portrait.jpg?rs=800',
        webpUrl: PROXY_BASE + '/images/portrait.jpg?f=webp&rs=800'
      },
      { 
        name: 'gif', 
        origSize: 1406568, // bytes (1.4 MB)
        webpUrl: PROXY_BASE + '/images/animated.gif?rs=600'
      }
    ];

    for (const comp of comparisons) {
      // Measure AVIF
      if (comp.avifUrl) {
        const avifSize = await measureImageSize(comp.avifUrl);
        const avifEl = document.getElementById('size-' + comp.name + '-avif');
        if (avifEl && avifSize) {
          avifEl.textContent = formatBytes(avifSize);
          avifEl.dataset.size = avifSize;
        }
      }

      // Measure WebP
      if (comp.webpUrl) {
        const webpSize = await measureImageSize(comp.webpUrl);
        const webpEl = document.getElementById('size-' + comp.name + '-webp');
        if (webpEl && webpSize) {
          webpEl.textContent = formatBytes(webpSize);
          webpEl.dataset.size = webpSize;
        }
      }

      // Calculate and display savings
      const savingsEl = document.getElementById('savings-' + comp.name);
      if (savingsEl) {
        const avifSize = parseInt(document.getElementById('size-' + comp.name + '-avif')?.dataset.size || '0', 10);
        const webpSize = parseInt(document.getElementById('size-' + comp.name + '-webp')?.dataset.size || '0', 10);
        
        let bestSize = avifSize || webpSize;
        if (avifSize && webpSize) {
          bestSize = Math.min(avifSize, webpSize);
        }

        if (bestSize > 0) {
          const savings = calculateSavings(comp.origSize, bestSize);
          if (savings !== null && savings > 0) {
            savingsEl.textContent = '💾 ' + savings + '% smaller than original (' + formatBytes(comp.origSize - bestSize) + ' saved)';
          } else if (savings !== null) {
            savingsEl.textContent = '⚠️ Similar to original size';
            savingsEl.style.color = 'var(--warning)';
          }
        }
      }
    }
  }

  // ========================================
  // Resize Gallery Sizes
  // ========================================

  async function updateResizeSizes() {
    const sizeElements = document.querySelectorAll('.resize-size[data-src]');
    
    for (const el of sizeElements) {
      const url = el.dataset.src;
      if (!url) continue;

      const size = await measureImageSize(url);
      if (size) {
        el.textContent = formatBytes(size);
        el.style.color = 'var(--success)';
      } else {
        el.textContent = 'Loading...';
      }
    }
  }

  // ========================================
  // Smooth scroll
  // ========================================

  function initSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach((anchor) => {
      anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
          target.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      });
    });
  }

  // ========================================
  // Initialize
  // ========================================

  document.addEventListener('DOMContentLoaded', () => {
    initSmoothScroll();
    
    // Wait a bit for images to start loading, then measure
    setTimeout(() => {
      updateFileComparisons();
      updateResizeSizes();
    }, 1000);

    // Retry after more images load
    setTimeout(() => {
      updateFileComparisons();
      updateResizeSizes();
    }, 5000);
  });
})();
