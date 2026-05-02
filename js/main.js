// DOM Elements
const header = document.querySelector('.header');
const hamburger = document.querySelector('.hamburger');
const navLinks = document.querySelector('.nav-links');
const navLinksItems = document.querySelectorAll('.nav-links a');
const slides = document.querySelectorAll('.slide');
const contactForm = document.getElementById('contact-form');
const authContainer = document.getElementById('auth-container');
const modal = document.getElementById('place-modal');
const closeModal = document.querySelector('.close-modal');
const modalBody = document.getElementById('modal-body');

// ============================================
// Intersection Observer for Animations
// ============================================
const revealObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            entry.target.classList.add('active');
            revealObserver.unobserve(entry.target);
        }
    });
}, { threshold: 0.15 });

document.querySelectorAll('.reveal').forEach(el => revealObserver.observe(el));

// ============================================
// Counter Animation with Intersection Observer
// ============================================
const statsSection = document.querySelector('.stats');
if (statsSection) {
    const statsObserver = new IntersectionObserver((entries) => {
        if (entries[0].isIntersecting) {
            const counters = statsSection.querySelectorAll('h3');
            counters.forEach(counter => {
                const target = counter.innerText;
                const num = parseInt(target.replace(/[^0-9]/g, ''));
                const suffix = target.replace(/[0-9,]/g, '');
                let current = 0;
                const duration = 2000;
                const increment = Math.ceil(num / (duration / 30));
                
                const timer = setInterval(() => {
                    current += increment;
                    if (current >= num) {
                        current = num;
                        clearInterval(timer);
                    }
                    counter.innerText = current.toLocaleString() + suffix;
                }, 30);
            });
            statsObserver.unobserve(statsSection);
        }
    }, { threshold: 0.2 });
    statsObserver.observe(statsSection);
}

// ============================================
// Parallax Effect on Hero (Optimized)
// ============================================
let ticking = false;
window.addEventListener('scroll', () => {
    if (!ticking) {
        window.requestAnimationFrame(() => {
            const scrolled = window.scrollY;
            const heroContent = document.querySelector('.hero-content');
            if (heroContent && scrolled < window.innerHeight) {
                heroContent.style.transform = `translateY(${scrolled * 0.3}px)`;
                heroContent.style.opacity = 1 - (scrolled / (window.innerHeight * 0.8));
            }
            ticking = false;
        });
        ticking = true;
    }
}, { passive: true });

// Load Site Content
async function loadSiteContent() {
    try {
        const res = await fetch('/api/content');
        if (res.ok) {
            const content = await res.json();
            const keys = ['hero_title', 'hero_subtitle', 'about_title', 'about_subtitle', 'about_text', 'contact_address', 'contact_phone', 'contact_email', 'destinations_title', 'destinations_subtitle'];
            
            keys.forEach(key => {
                const el = document.getElementById(key.replace('_', '-'));
                if (el && content[key]) {
                    el.innerText = content[key];
                }
            });

            for(let i=1; i<=4; i++) {
                const nameEl = document.getElementById(`dest${i}-name`);
                const descEl = document.getElementById(`dest${i}-desc`);
                const imgEl = document.getElementById(`dest${i}-img`);
                if(nameEl && content[`dest${i}_name`]) nameEl.innerText = content[`dest${i}_name`];
                if(descEl && content[`dest${i}_desc`]) descEl.innerText = content[`dest${i}_desc`];
                if(imgEl && content[`dest${i}_img`]) imgEl.src = content[`dest${i}_img`];
            }

            for(let i=1; i<=3; i++) {
                const heroImgEl = document.getElementById(`hero-img${i}`);
                if(heroImgEl && content[`hero_img${i}`]) {
                    heroImgEl.style.backgroundImage = `url('${content[`hero_img${i}`]}')`;
                }
            }
        }
    } catch (err) {
        console.error("Failed to load site content:", err);
    }
}
loadSiteContent();

// Check User Login Status
async function checkAuth() {
    try {
        const res = await fetch('/api/user');
        if (res.ok) {
            const data = await res.json();
            authContainer.innerHTML = `
                <div style="color: white; font-weight: bold; margin-left: 20px; display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-user-circle"></i> ${data.name || 'Foydalanuvchi'}
                    <a href="/auth/google/logout" class="btn-secondary" style="color:white; border-color:white; padding: 5px 10px; font-size: 0.8rem; margin-left: 10px;">Chiqish</a>
                </div>
            `;
            
            const nameInput = document.getElementById('name');
            const emailInput = document.getElementById('email');
            if(nameInput && data.name) nameInput.value = data.name;
            if(emailInput && data.email) {
                emailInput.value = data.email;
                emailInput.setAttribute('readonly', 'readonly');
                emailInput.style.backgroundColor = '#f0f0f0';
                emailInput.style.cursor = 'not-allowed';
            }
        }
    } catch (err) {
        // Silent error for auth
    }
}
checkAuth();

// Header scroll effect
window.addEventListener('scroll', () => {
    if (window.scrollY > 100) {
        header.classList.add('scrolled');
    } else {
        header.classList.remove('scrolled');
    }
}, { passive: true });

// Mobile menu toggle
hamburger.addEventListener('click', () => {
    const expanded = hamburger.getAttribute('aria-expanded') === 'true';
    hamburger.setAttribute('aria-expanded', !expanded);
    navLinks.classList.toggle('active');
    hamburger.classList.toggle('active');
});

navLinksItems.forEach(link => {
    link.addEventListener('click', () => {
        navLinks.classList.remove('active');
        hamburger.classList.remove('active');
        hamburger.setAttribute('aria-expanded', 'false');
    });
});

// Hero slider
let currentSlide = 0;
function nextSlide() {
    if (slides.length === 0) return;
    slides[currentSlide].classList.remove('active');
    currentSlide = (currentSlide + 1) % slides.length;
    slides[currentSlide].classList.add('active');
}
if (slides.length > 0) setInterval(nextSlide, 5000);

// Load places from API
async function loadPlaces() {
    try {
        const res = await fetch('/api/places');
        if (!res.ok) throw new Error("Failed to fetch places");
        const places = await res.json();
        displayPlaces(places);
    } catch (err) {
        console.error(err);
        document.getElementById('places-container').innerHTML = '<p class="no-places">Joylar yuklanmadi. Server ishlayotganiga ishonch hosil qiling.</p>';
    }
}

function displayPlaces(places) {
    const placesContainer = document.getElementById('places-container');
    
    if (!places || places.length === 0) {
        placesContainer.innerHTML = '<p class="no-places">Hozircha turistik joylar yo\'q</p>';
        return;
    }
    
    placesContainer.innerHTML = places.map(item => `
        <div class="place-card" style="cursor: pointer;" onclick="openPlaceDetails(${item.id})">
            <div class="place-image">
                <img src="${item.images && item.images.length > 0 ? item.images[0] : 'https://via.placeholder.com/600x400?text=Joy'}" alt="${item.name}" loading="lazy" width="600" height="400">
                <span class="place-badge">${item.category}</span>
            </div>
            <div class="place-content">
                <div class="place-location">
                    <i class="fas fa-map-marker-alt"></i>
                    <span>${item.location}</span>
                </div>
                <h3>${item.name}</h3>
                <p class="description">${item.description.substring(0, 100)}...</p>
                <div style="margin-top: 10px; font-weight: bold; color: var(--primary);">Batafsil ma'lumot va narxlar <i class="fas fa-arrow-right"></i></div>
            </div>
        </div>
    `).join('');
}

// Open Place Details Modal
async function openPlaceDetails(id) {
    try {
        const res = await fetch(`/api/places/${id}`);
        if (!res.ok) throw new Error("Failed to fetch details");
        const place = await res.json();
        
        let imagesHtml = '';
        if (place.images && place.images.length > 0) {
            imagesHtml = `
                <div class="carousel-container">
                    ${place.images.map((img, idx) => `<img class="carousel-slide ${idx===0?'active':''}" src="${img}" alt="${place.name}" loading="lazy">`).join('')}
                    ${place.images.length > 1 ? `
                        <a class="carousel-prev" onclick="moveSlide(-1)" aria-label="Oldingi rasm">&#10094;</a>
                        <a class="carousel-next" onclick="moveSlide(1)" aria-label="Keyingi rasm">&#10095;</a>
                    ` : ''}
                </div>
            `;
        }
        
        modalBody.innerHTML = `
            <h2 style="margin-bottom: 10px;">${place.name}</h2>
            <div style="color: #666; margin-bottom: 15px;"><i class="fas fa-map-marker-alt"></i> ${place.location} &bull; ${place.category}</div>
            ${imagesHtml}
            <p style="font-size: 1.1rem; line-height: 1.6; margin-bottom: 20px;">${place.description}</p>
            <div class="modal-price" style="margin-bottom: 20px;">Sayohat narxi: ${place.price.toLocaleString()} so'm</div>
            <div style="text-align: center;">
                <a href="https://t.me/+998913328290" target="_blank" class="btn-primary" style="display:inline-block; padding:10px 20px; border-radius:5px; background: #0088cc; color:#fff; text-decoration:none;"><i class="fab fa-telegram"></i> Onlayn bron qilish</a>
            </div>
        `;
        
        modal.style.display = "block";
        window.currentCarouselIndex = 0;
    } catch (err) {
        console.error(err);
        alert("Ma'lumotlarni yuklashda xatolik yuz berdi.");
    }
}

// Open Destination Details Modal
window.openDestinationModal = function(destId) {
    const name = document.getElementById(`dest${destId}-name`).innerText;
    const desc = document.getElementById(`dest${destId}-desc`).innerText;
    const img = document.getElementById(`dest${destId}-img`).src;
    
    modalBody.innerHTML = `
        <h2 style="margin-bottom: 10px;">${name}</h2>
        <div style="color: #666; margin-bottom: 15px;"><i class="fas fa-map-marker-alt"></i> O'zbekiston &bull; Mashhur Yo'nalish</div>
        <div class="carousel-container">
            <img class="carousel-slide active" src="${img}" alt="${name}" loading="lazy">
        </div>
        <p style="font-size: 1.1rem; line-height: 1.6; margin-bottom: 20px;">${desc}</p>
        <div style="margin-top: 20px; text-align:center; display:flex; gap:10px; justify-content:center; flex-wrap:wrap;">
            <a href="#places" onclick="modal.style.display = 'none'" class="btn-primary" style="display:inline-block; padding:10px 20px; border-radius:5px; color:#fff; text-decoration:none;">Sayohat paketlarini ko'rish</a>
            <a href="https://t.me/+998913328290" target="_blank" class="btn-primary" style="display:inline-block; padding:10px 20px; border-radius:5px; background: #0088cc; color:#fff; text-decoration:none;"><i class="fab fa-telegram"></i> Onlayn bron qilish</a>
        </div>
    `;
    
    modal.style.display = "block";
}

window.moveSlide = function(step) {
    const slides = document.querySelectorAll('.carousel-slide');
    if(slides.length === 0) return;
    slides[window.currentCarouselIndex].classList.remove('active');
    window.currentCarouselIndex = (window.currentCarouselIndex + step + slides.length) % slides.length;
    slides[window.currentCarouselIndex].classList.add('active');
}

closeModal.onclick = function() {
    modal.style.display = "none";
}
window.onclick = function(event) {
    if (event.target == modal) {
        modal.style.display = "none";
    }
}

// Contact form submission
contactForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const message = document.getElementById('message').value;
    
    try {
        const res = await fetch('/api/contact', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ name, email, message })
        });
        
        if (res.ok) {
            alert(`Rahmat ${name}! Xabaringiz qabul qilindi. Tez orada siz bilan bog'lanamiz.`);
            const emailInput = document.getElementById('email');
            if (emailInput.hasAttribute('readonly')) {
                document.getElementById('message').value = '';
            } else {
                contactForm.reset();
            }
        } else {
            alert("Xatolik yuz berdi. Iltimos qayta urinib ko'ring.");
        }
    } catch (err) {
        console.error("Submit error:", err);
        alert("Xatolik yuz berdi. Iltimos qayta urinib ko'ring.");
    }
});

// Initialize
loadPlaces();

// Smooth scroll for anchor links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            target.scrollIntoView({
                behavior: 'smooth',
                block: 'start'
            });
        }
    });
});
