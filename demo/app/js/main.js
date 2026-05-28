(function () {
  'use strict';

  document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('a[href^="#"]').forEach((anchor) => {
      anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
          target.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      });
    });

    initPlayground();
  });

  function initPlayground() {
    const data = {
      'photo.jpg': {
        orig: '4.7 MB',
        origBytes: 4.7 * 1024 * 1024,
        sizes: {
          400:  { label: '12.6 KB', bytes: 12.6 * 1024 },
          800:  { label: '41.8 KB', bytes: 41.8 * 1024 },
          1200: { label: '90.8 KB', bytes: 90.8 * 1024 },
          1920: { label: '216.4 KB', bytes: 216.4 * 1024 },
        },
      },
      'portrait.jpg': {
        orig: '3.3 MB',
        origBytes: 3.3 * 1024 * 1024,
        sizes: {
          400:  { label: '5.3 KB',  bytes: 5.3 * 1024 },
          800:  { label: '12.7 KB', bytes: 12.7 * 1024 },
          1200: { label: '21.6 KB', bytes: 21.6 * 1024 },
          1920: { label: '44.3 KB', bytes: 44.3 * 1024 },
        },
      },
    };

    const MAX_WIDTH = 1920;
    const state = { image: 'photo.jpg', size: 800, quality: '', blur: '' };

    const img = document.getElementById('playground-img');
    const urlEl = document.getElementById('playground-url');
    const barFill = document.getElementById('size-bar-fill');
    const barLabel = document.getElementById('size-bar-label');
    const statOrig = document.getElementById('stat-orig');
    const statOpt = document.getElementById('stat-opt');
    const statSavings = document.getElementById('stat-savings');

    if (!img) return;

    function update() {
      const { image, size, quality, blur } = state;
      const imageData = data[image];

      const params = ['rs=' + size];
      if (quality) params.push('q=' + quality);
      if (blur) params.push('blur=' + blur);
      const src = 'images/' + image + '?' + params.join('&');

      img.style.opacity = '0.5';
      img.onload = () => { img.style.opacity = '1'; };
      img.src = src;
      img.style.width = size + 'px';

      urlEl.textContent = src;

      const pct = Math.round(size / MAX_WIDTH * 100);
      barFill.style.width = pct + '%';
      barLabel.textContent = size + 'px wide · ' + pct + '% of 1920px';

      const sizeInfo = imageData.sizes[size];
      const hasEffects = quality || blur;

      statOrig.textContent = imageData.orig;
      statOpt.textContent = sizeInfo.label + (hasEffects ? '*' : '');

      const reduction = Math.round((1 - sizeInfo.bytes / imageData.origBytes) * 100);
      statSavings.textContent = (hasEffects ? '~' : '') + reduction + '%' + (hasEffects ? '+' : '');
    }

    const groups = [
      { id: 'pills-image',   key: 'image',   parse: (v) => v },
      { id: 'pills-size',    key: 'size',    parse: (v) => parseInt(v, 10) },
      { id: 'pills-quality', key: 'quality', parse: (v) => v },
      { id: 'pills-blur',    key: 'blur',    parse: (v) => v },
    ];

    groups.forEach(({ id, key, parse }) => {
      const group = document.getElementById(id);
      if (!group) return;
      group.querySelectorAll('.pill').forEach((pill) => {
        pill.addEventListener('click', () => {
          group.querySelectorAll('.pill').forEach((p) => p.classList.remove('active'));
          pill.classList.add('active');
          state[key] = parse(pill.dataset.value);
          update();
        });
      });
    });

    update();
  }
})();
