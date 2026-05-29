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
    // combos[width][quality][blur] — all values measured from the real proxy
    const data = {
      'photo.jpg': {
        orig: '4.7 MB',
        origBytes: 4928307,
        combos: {
          400: {
            auto: { 0: { label: '12.6 KB', bytes: 12863 }, 2: { label: '6.1 KB',  bytes: 6212  }, 5: { label: '4.3 KB', bytes: 4416 }, 10: { label: '3.4 KB', bytes: 3496 } },
            90:   { 0: { label: '32.3 KB', bytes: 33114 }, 2: { label: '12.9 KB', bytes: 13216 }, 5: { label: '9.1 KB', bytes: 9326 }, 10: { label: '7.3 KB', bytes: 7439 } },
            75:   { 0: { label: '18.2 KB', bytes: 18599 }, 2: { label: '7.9 KB',  bytes: 8114  }, 5: { label: '5.6 KB', bytes: 5725 }, 10: { label: '4.4 KB', bytes: 4549 } },
            50:   { 0: { label: '8.2 KB',  bytes: 8410  }, 2: { label: '4.6 KB',  bytes: 4679  }, 5: { label: '3.4 KB', bytes: 3522 }, 10: { label: '2.7 KB', bytes: 2810 } },
          },
          800: {
            auto: { 0: { label: '41.8 KB',  bytes: 42829  }, 2: { label: '18.1 KB', bytes: 18532 }, 5: { label: '11.5 KB', bytes: 11805 }, 10: { label: '8.8 KB',  bytes: 8978  } },
            90:   { 0: { label: '113.9 KB', bytes: 116639 }, 2: { label: '41.6 KB', bytes: 42614 }, 5: { label: '26.6 KB', bytes: 27217 }, 10: { label: '21.1 KB', bytes: 21609 } },
            75:   { 0: { label: '61.9 KB',  bytes: 63427  }, 2: { label: '24.4 KB', bytes: 24960 }, 5: { label: '15.4 KB', bytes: 15728 }, 10: { label: '11.9 KB', bytes: 12149 } },
            50:   { 0: { label: '26.5 KB',  bytes: 27187  }, 2: { label: '13.2 KB', bytes: 13498 }, 5: { label: '8.9 KB',  bytes: 9125  }, 10: { label: '6.7 KB',  bytes: 6893  } },
          },
          1200: {
            auto: { 0: { label: '90.8 KB',  bytes: 92961  }, 2: { label: '36.6 KB', bytes: 37517 }, 5: { label: '21.2 KB', bytes: 21704 }, 10: { label: '15.0 KB', bytes: 15395 } },
            90:   { 0: { label: '253.8 KB', bytes: 259888 }, 2: { label: '86.9 KB', bytes: 89017 }, 5: { label: '51.4 KB', bytes: 52679 }, 10: { label: '38.4 KB', bytes: 39324 } },
            75:   { 0: { label: '135.8 KB', bytes: 139030 }, 2: { label: '50.1 KB', bytes: 51274 }, 5: { label: '28.7 KB', bytes: 29417 }, 10: { label: '20.9 KB', bytes: 21356 } },
            50:   { 0: { label: '57.1 KB',  bytes: 58420  }, 2: { label: '26.3 KB', bytes: 26946 }, 5: { label: '15.9 KB', bytes: 16316 }, 10: { label: '11.3 KB', bytes: 11613 } },
          },
          1920: {
            auto: { 0: { label: '216.4 KB', bytes: 221552 }, 2: { label: '86.5 KB',  bytes: 88589  }, 5: { label: '46.1 KB',  bytes: 47232  }, 10: { label: '30.7 KB', bytes: 31476 } },
            90:   { 0: { label: '636.4 KB', bytes: 651630 }, 2: { label: '209.4 KB', bytes: 214437 }, 5: { label: '116.8 KB', bytes: 119631 }, 10: { label: '82.7 KB', bytes: 84680 } },
            75:   { 0: { label: '328.9 KB', bytes: 336782 }, 2: { label: '118.9 KB', bytes: 121726 }, 5: { label: '63.9 KB',  bytes: 65460  }, 10: { label: '43.3 KB', bytes: 44362 } },
            50:   { 0: { label: '134.7 KB', bytes: 137893 }, 2: { label: '61.2 KB',  bytes: 62655  }, 5: { label: '34.2 KB',  bytes: 35005  }, 10: { label: '22.9 KB', bytes: 23479 } },
          },
        },
      },
      'portrait.jpg': {
        orig: '3.3 MB',
        origBytes: 3460300,
        combos: {
          400: {
            auto: { 0: { label: '5.3 KB',  bytes: 5457  }, 2: { label: '3.9 KB', bytes: 4039 }, 5: { label: '3.3 KB', bytes: 3350 }, 10: { label: '2.7 KB', bytes: 2757 } },
            90:   { 0: { label: '12.5 KB', bytes: 12772 }, 2: { label: '8.3 KB', bytes: 8499 }, 5: { label: '6.8 KB', bytes: 6981 }, 10: { label: '5.7 KB', bytes: 5835 } },
            75:   { 0: { label: '7.3 KB',  bytes: 7457  }, 2: { label: '5.1 KB', bytes: 5182 }, 5: { label: '4.2 KB', bytes: 4257 }, 10: { label: '3.5 KB', bytes: 3553 } },
            50:   { 0: { label: '3.9 KB',  bytes: 4019  }, 2: { label: '3.1 KB', bytes: 3185 }, 5: { label: '2.6 KB', bytes: 2685 }, 10: { label: '2.2 KB', bytes: 2264 } },
          },
          800: {
            auto: { 0: { label: '12.7 KB', bytes: 13013 }, 2: { label: '9.1 KB',  bytes: 9276  }, 5: { label: '7.4 KB',  bytes: 7560  }, 10: { label: '6.2 KB',  bytes: 6388  } },
            90:   { 0: { label: '33.3 KB', bytes: 34129 }, 2: { label: '21.4 KB', bytes: 21923 }, 5: { label: '17.5 KB', bytes: 17909 }, 10: { label: '15.4 KB', bytes: 15808 } },
            75:   { 0: { label: '17.9 KB', bytes: 18372 }, 2: { label: '12.2 KB', bytes: 12456 }, 5: { label: '9.8 KB',  bytes: 10021 }, 10: { label: '8.5 KB',  bytes: 8659  } },
            50:   { 0: { label: '9.0 KB',  bytes: 9254  }, 2: { label: '6.8 KB',  bytes: 6971  }, 5: { label: '5.7 KB',  bytes: 5809  }, 10: { label: '4.8 KB',  bytes: 4964  } },
          },
          1200: {
            auto: { 0: { label: '21.6 KB', bytes: 22145 }, 2: { label: '14.8 KB', bytes: 15203 }, 5: { label: '11.9 KB', bytes: 12140 }, 10: { label: '10.0 KB', bytes: 10236 } },
            90:   { 0: { label: '61.4 KB', bytes: 62864 }, 2: { label: '38.1 KB', bytes: 39010 }, 5: { label: '30.9 KB', bytes: 31659 }, 10: { label: '27.2 KB', bytes: 27838 } },
            75:   { 0: { label: '31.5 KB', bytes: 32226 }, 2: { label: '20.5 KB', bytes: 20997 }, 5: { label: '16.4 KB', bytes: 16814 }, 10: { label: '13.9 KB', bytes: 14210 } },
            50:   { 0: { label: '14.9 KB', bytes: 15247 }, 2: { label: '10.9 KB', bytes: 11172 }, 5: { label: '8.9 KB',  bytes: 9148  }, 10: { label: '7.5 KB',  bytes: 7707  } },
          },
          1920: {
            auto: { 0: { label: '44.2 KB',  bytes: 45308  }, 2: { label: '28.3 KB', bytes: 28946 }, 5: { label: '22.0 KB', bytes: 22494 }, 10: { label: '18.5 KB', bytes: 18940 } },
            90:   { 0: { label: '141.1 KB', bytes: 144459 }, 2: { label: '78.1 KB', bytes: 79918 }, 5: { label: '61.9 KB', bytes: 63364 }, 10: { label: '54.9 KB', bytes: 56167 } },
            75:   { 0: { label: '65.9 KB',  bytes: 67430  }, 2: { label: '40.0 KB', bytes: 40984 }, 5: { label: '31.0 KB', bytes: 31770 }, 10: { label: '26.4 KB', bytes: 27036 } },
            50:   { 0: { label: '29.9 KB',  bytes: 30663  }, 2: { label: '20.4 KB', bytes: 20938 }, 5: { label: '16.3 KB', bytes: 16680 }, 10: { label: '13.8 KB', bytes: 14118 } },
          },
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

    function formatBytes(bytes) {
      const kb = bytes / 1024;
      return (kb >= 1000 ? Math.round(kb).toLocaleString('en') : kb.toFixed(1)) + ' KB';
    }

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

      const qKey = quality || 'auto';
      const bKey = parseInt(blur || '0', 10);
      const sizeInfo = imageData.combos[size][qKey][bKey];

      statOrig.textContent = formatBytes(imageData.origBytes);
      statOpt.textContent = sizeInfo.label;

      const savedBytes = imageData.origBytes - sizeInfo.bytes;
      statSavings.textContent = formatBytes(savedBytes);
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
