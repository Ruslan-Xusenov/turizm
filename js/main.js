// DOM Elements
const header = document.querySelector('.header');
const hamburger = document.querySelector('.hamburger');
const navLinks = document.querySelector('.nav-links');
const navLinksItems = document.querySelectorAll('.nav-links a');
const slides = document.querySelectorAll('.slide');
const contactForm = document.getElementById('contact-form');

// Header scroll effect
window.addEventListener('scroll', () => {
    if (window.scrollY > 100) {
        header.classList.add('scrolled');
    } else {
        header.classList.remove('scrolled');
    }
});

// Mobile menu toggle
hamburger.addEventListener('click', () => {
    navLinks.classList.toggle('active');
});

// Close mobile menu on link click
navLinksItems.forEach(link => {
    link.addEventListener('click', () => {
        navLinks.classList.remove('active');
    });
});

// Hero slider
let currentSlide = 0;

function nextSlide() {
    slides[currentSlide].classList.remove('active');
    currentSlide = (currentSlide + 1) % slides.length;
    slides[currentSlide].classList.add('active');
}

setInterval(nextSlide, 5000);

// Active navigation link on scroll
const sections = document.querySelectorAll('section[id]');

window.addEventListener('scroll', () => {
    let current = '';
    
    sections.forEach(section => {
        const sectionTop = section.offsetTop;
        const sectionHeight = section.clientHeight;
        
        if (scrollY >= sectionTop - 200) {
            current = section.getAttribute('id');
        }
    });
    
    navLinksItems.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === `#${current}`) {
            link.classList.add('active');
        }
    });
});

// Load places from localStorage
function loadPlaces() {
    const placesContainer = document.getElementById('places-container');
    const places = JSON.parse(localStorage.getItem('places')) || [];
    
    if (places.length === 0) {
        // Show default places if no places in localStorage
        const defaultPlaces = [
            {
                id: 1,
                name: 'Registon maydoni',
                location: 'Samarqand',
                description: 'Samarqandning eng mashhur diqqatga sazovor joyi. Uch madrasa - Ulug\'bek, Sherdor va Tillakori madrasalaridan iborat. XV asrda qurilgan bu majmua dunyoviy me\'morlikning eng yaxshi namunalaridan biri hisoblanadi.',
                image: 'https://images.unsplash.com/photo-1564507592333-c60657eea523?w=600',
                highlights: ['XV asr me\'morligi', 'Uch madrasa majmuasi', 'Kunduz va tun ko\'rinishi'],
                category: 'Arxitektura'
            },
            {
                id: 2,
                name: 'Go\'ri Amiq',
                location: 'Samarqand',
                description: 'Amir Temurning maqbarasi. Zardushtiylar dinining ta\'siri ostida qurilgan bu inshoot katta sharqiy xazinalar va yuqori sifatli naqshlar bilan mashhur.',
                image: 'https://images.unsplash.com/photo-1583417319070-4a69db38a482?w=600',
                highlights: ['Amir Temur maqbarasi', 'Zumrad mozaikalar', 'Tarixiy muzey'],
                category: 'Tarixiy obida'
            },
            {
                id: 3,
                name: 'Ichan-Qal\'a',
                location: 'Xiva',
                description: 'UNESCO tomonidan dunyoviy meros sifatida tan olingan Xivaning ichki shaharchasi. 2500 yillik tarixga ega bo\'lgan bu joyda 50 dan ortiq tarixiy obidalar saqlanib qolgan.',
                image: 'https://images.unsplash.com/photo-1599576935803-91efed66198f?w=600',
                highlights: ['UNESCO merosi', '2500 yillik tarix', '50+ tarixiy obida'],
                category: 'Qadimiy shahar'
            },
            {
                id: 4,
                name: 'Lyabi-Hovuz',
                location: 'Buxoro',
                description: 'Buxoroning markaziy maydoni va eng mashhur diqqatga sazovor joyi. XVII asrda qurilgan hovuz atrofida Nodir Devonbegi madrasasi, Xonako va Mag\'oki-Attori masjidi joylashgan.',
                image: 'https://images.unsplash.com/photo-1548013146-72479768bada?w=600',
                highlights: ['Markaziy hovuz', 'Madrasa va xonako', 'Kechki yoritish'],
                category: 'Majmua'
            },
            {
                id: 5,
                name: 'Chorvoq',
                location: 'Toshkent viloyati',
                description: 'Chatqol tog\'lari etaklarida joylashgan sun\'iy ko\'l va dam olish zonasi. Tabiiy go\'zallik va toza havo qidiruvchilar uchun ideal joy.',
                image: 'https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=600',
                highlights: ['Tog\' ko\'li', 'Dam olish zonalari', 'Sport turlari'],
                category: 'Tabiat'
            },
            {
                id: 6,
                name: 'Shahrisabz',
                location: 'Qashqadaryo',
                description: 'Amir Temurning vatani va dunyoviy meros ro\'yxatiga kiritilgan shahar. Ak-Saroy saroyi, Dor-ut Tilovat va Dor-us Saodat majmualari bilan mashhur.',
                image: 'https://images.unsplash.com/photo-1563284223-6233e93d7dc2?w=600',
                highlights: ['Ak-Saroy saroyi', 'Amir Temur vatani', 'UNESCO merosi'],
                category: 'Tarixiy shahar'
            }
        ];
        
        localStorage.setItem('places', JSON.stringify(defaultPlaces));
        displayPlaces(defaultPlaces);
    } else {
        displayPlaces(places);
    }
}

function displayPlaces(places) {
    const placesContainer = document.getElementById('places-container');
    
    if (places.length === 0) {
        placesContainer.innerHTML = '<p class="no-places">Hozircha turistik joylar yo\'q</p>';
        return;
    }
    
    placesContainer.innerHTML = places.map(item => `
        <div class="place-card">
            <div class="place-image">
                <img src="${item.image}" alt="${item.name}" onerror="this.src='https://via.placeholder.com/600x400?text=Place'">
                <span class="place-badge">${item.category}</span>
            </div>
            <div class="place-content">
                <div class="place-location">
                    <i class="fas fa-map-marker-alt"></i>
                    <span>${item.location}</span>
                </div>
                <h3>${item.name}</h3>
                <p class="description">${item.description}</p>
                <div class="place-details">
                    <h4>Diqqatga sazovor:</h4>
                    <ul>
                        ${item.highlights.map(h => `<li><i class="fas fa-star"></i> ${h}</li>`).join('')}
                    </ul>
                </div>
            </div>
        </div>
    `).join('');
}

// Contact form submission
contactForm.addEventListener('submit', (e) => {
    e.preventDefault();
    
    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const message = document.getElementById('message').value;
    
    // In a real application, you would send this data to a server
    // For now, we'll just show a success message
    alert(`Rahmat ${name}! Xabaringiz qabul qilindi. Tez orada siz bilan bog'lanamiz.`);
    
    contactForm.reset();
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
