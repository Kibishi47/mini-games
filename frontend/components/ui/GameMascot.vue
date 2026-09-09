<template>
  <div :class="['inline-block select-none relative', sizeClass, animationClass]">
    <svg
      viewBox="0 0 100 100"
      class="w-full h-full overflow-visible"
      xmlns="http://www.w3.org/2000/svg"
    >
      <!-- Ombre portée / Hachures sous les pieds (||||) -->
      <g v-if="mood !== 'dead'">
        <ellipse cx="50" cy="94" rx="22" ry="4" fill="#121212" opacity="0.25" />
        <line x1="38" y1="94" x2="42" y2="98" stroke="#121212" stroke-width="2" />
        <line x1="48" y1="94" x2="52" y2="98" stroke="#121212" stroke-width="2" />
        <line x1="58" y1="94" x2="62" y2="98" stroke="#121212" stroke-width="2" />
      </g>

      <!-- Étoiles de fête si Happy -->
      <g v-if="mood === 'happy'">
        <polygon points="12,18 15,24 22,24 16,28 18,34 12,30 6,34 8,28 2,24 9,24" fill="#FFD300" stroke="#121212" stroke-width="2" />
        <polygon points="86,16 88,20 93,20 89,23 90,27 86,24 82,27 83,23 79,20 84,20" fill="#FFD300" stroke="#121212" stroke-width="2" />
      </g>

      <!-- Lettres ZZZ si Sleeping -->
      <g v-if="mood === 'sleeping'" class="font-display font-black text-ink-black text-xs">
        <text x="75" y="25" fill="#121212" font-size="14" font-weight="900">Z</text>
        <text x="84" y="16" fill="#121212" font-size="11" font-weight="900">z</text>
      </g>

      <!-- ================= 1. DICE (Le Dé Sprinteur) ================= -->
      <g v-if="name === 'dice'">
        <!-- Jambes rubberhose -->
        <path d="M38 72 C 34 85, 26 86, 22 92" stroke="#E63228" stroke-width="6" stroke-linecap="round" fill="none" />
        <path d="M38 72 C 34 85, 26 86, 22 92" stroke="#121212" stroke-width="1.5" stroke-linecap="round" fill="none" />
        <ellipse cx="20" cy="92" rx="6" ry="3" fill="#121212" />

        <path d="M62 72 C 66 84, 74 86, 78 92" stroke="#E63228" stroke-width="6" stroke-linecap="round" fill="none" />
        <ellipse cx="80" cy="92" rx="6" ry="3" fill="#121212" />

        <!-- Cube Blanc Dé -->
        <rect x="22" y="22" width="56" height="54" rx="10" fill="#FFFFFF" stroke="#121212" stroke-width="4" />
        <!-- Points de Dé (Pips) -->
        <circle cx="34" cy="34" r="4.5" fill="#121212" />
        <circle cx="66" cy="34" r="4.5" fill="#121212" />
        <circle cx="34" cy="64" r="4.5" fill="#121212" />
        <circle cx="66" cy="64" r="4.5" fill="#121212" />

        <!-- Visage du Dé -->
        <g v-if="mood === 'dead'">
          <line x1="42" y1="45" x2="48" y2="51" stroke="#121212" stroke-width="3" stroke-linecap="round" />
          <line x1="48" y1="45" x2="42" y2="51" stroke="#121212" stroke-width="3" stroke-linecap="round" />
          <line x1="52" y1="45" x2="58" y2="51" stroke="#121212" stroke-width="3" stroke-linecap="round" />
          <line x1="58" y1="45" x2="52" y2="51" stroke="#121212" stroke-width="3" stroke-linecap="round" />
          <line x1="44" y1="58" x2="56" y2="58" stroke="#121212" stroke-width="3" stroke-linecap="round" />
        </g>
        <g v-else>
          <!-- Yeux ronds expressifs -->
          <circle cx="44" cy="46" r="5" fill="#121212" />
          <circle cx="43" cy="45" r="1.5" fill="#FFFFFF" />
          <circle cx="56" cy="46" r="5" fill="#121212" />
          <circle cx="55" cy="45" r="1.5" fill="#FFFFFF" />
          <!-- Sourire cartoon -->
          <path v-if="mood === 'happy'" d="M42 54 Q 50 64, 58 54" stroke="#121212" stroke-width="3" stroke-linecap="round" fill="#E63228" />
          <path v-else-if="mood === 'panic'" d="M44 56 Q 50 51, 56 56" stroke="#121212" stroke-width="3" stroke-linecap="round" fill="none" />
          <path v-else d="M44 55 Q 50 60, 56 55" stroke="#121212" stroke-width="3" stroke-linecap="round" fill="none" />
        </g>
      </g>

      <!-- ================= 2. DOMINO (Le Flegmatique) ================= -->
      <g v-else-if="name === 'domino'">
        <!-- Jambes -->
        <path d="M42 78 L 38 92" stroke="#121212" stroke-width="4" stroke-linecap="round" />
        <ellipse cx="36" cy="92" rx="5" ry="3" fill="#121212" />
        <path d="M58 78 L 62 92" stroke="#121212" stroke-width="4" stroke-linecap="round" />
        <ellipse cx="64" cy="92" rx="5" ry="3" fill="#121212" />

        <!-- Corps Domino -->
        <rect x="30" y="16" width="40" height="64" rx="8" fill="#FFFFFF" stroke="#121212" stroke-width="4" />
        <line x1="32" y1="48" x2="68" y2="48" stroke="#121212" stroke-width="3" />
        <!-- Pips bas -->
        <circle cx="40" cy="58" r="3.5" fill="#121212" />
        <circle cx="60" cy="58" r="3.5" fill="#121212" />
        <circle cx="50" cy="68" r="3.5" fill="#121212" />

        <!-- Visage haut -->
        <g v-if="mood === 'dead'">
          <text x="36" y="38" font-size="12" font-weight="900" fill="#121212">X</text>
          <text x="52" y="38" font-size="12" font-weight="900" fill="#121212">X</text>
        </g>
        <g v-else>
          <circle cx="42" cy="32" r="5" fill="#121212" />
          <circle cx="41" cy="31" r="1.5" fill="#FFFFFF" />
          <circle cx="58" cy="32" r="5" fill="#121212" />
          <circle cx="57" cy="31" r="1.5" fill="#FFFFFF" />
          <line x1="46" y1="40" x2="54" y2="40" stroke="#121212" stroke-width="2.5" stroke-linecap="round" />
        </g>
      </g>

      <!-- ================= 3. CARD (L'As de Cœur) ================= -->
      <g v-else-if="name === 'card'">
        <!-- Bras gantés & Jambes bleues -->
        <path d="M26 50 C 14 55, 12 40, 6 45" stroke="#1D4ED8" stroke-width="4" stroke-linecap="round" fill="none" />
        <circle cx="6" cy="45" r="4" fill="#FFFFFF" stroke="#121212" stroke-width="2" />
        <path d="M74 50 C 86 55, 88 40, 94 45" stroke="#1D4ED8" stroke-width="4" stroke-linecap="round" fill="none" />
        <circle cx="94" cy="45" r="4" fill="#FFFFFF" stroke="#121212" stroke-width="2" />

        <path d="M40 76 L 36 92" stroke="#1D4ED8" stroke-width="5" stroke-linecap="round" />
        <ellipse cx="34" cy="92" rx="5" ry="3" fill="#121212" />
        <path d="M60 76 L 64 92" stroke="#1D4ED8" stroke-width="5" stroke-linecap="round" />
        <ellipse cx="66" cy="92" rx="5" ry="3" fill="#121212" />

        <!-- Carte -->
        <rect x="25" y="16" width="50" height="62" rx="6" fill="#FFFFFF" stroke="#121212" stroke-width="4" />
        <text x="30" y="30" font-family="Archivo Black" font-size="10" fill="#E63228">A</text>
        <!-- Cœur central -->
        <path d="M50 42 C 50 35, 40 33, 40 40 C 40 47, 50 54, 50 54 C 50 54, 60 47, 60 40 C 60 33, 50 35, 50 42 Z" fill="#E63228" />

        <!-- Visage -->
        <circle cx="43" cy="34" r="3.5" fill="#121212" />
        <circle cx="57" cy="34" r="3.5" fill="#121212" />
        <path d="M46 62 Q 50 67, 54 62" stroke="#121212" stroke-width="2.5" stroke-linecap="round" fill="none" />
      </g>

      <!-- ================= 4. KNIGHT (Le Cavalier) ================= -->
      <g v-else-if="name === 'knight'">
        <!-- Sabots / Jambes -->
        <rect x="36" y="80" width="8" height="12" fill="#121212" rx="2" />
        <rect x="56" y="80" width="8" height="12" fill="#121212" rx="2" />

        <!-- Tête de cheval échecs -->
        <path
          d="M34 80 L 34 50 C 30 46, 22 46, 22 36 C 22 28, 30 24, 38 24 C 42 16, 52 14, 62 16 C 70 18, 76 26, 76 40 C 76 56, 68 70, 68 80 Z"
          fill="#E63228"
          stroke="#121212"
          stroke-width="4"
          stroke-linejoin="round"
        />
        <circle cx="48" cy="30" r="4.5" fill="#FFFFFF" stroke="#121212" stroke-width="2" />
        <circle cx="49" cy="30" r="2" fill="#121212" />
        <!-- Crinière stylisée -->
        <polygon points="68,26 76,28 72,34" fill="#121212" />
        <polygon points="70,38 78,40 73,46" fill="#121212" />
        <!-- Museau narine -->
        <circle cx="28" cy="36" r="2" fill="#121212" />
      </g>

      <!-- ================= 5. D20 (Polyèdre) ================= -->
      <g v-else-if="name === 'd20'">
        <!-- Gambettes -->
        <path d="M40 76 L 36 92" stroke="#121212" stroke-width="4" stroke-linecap="round" />
        <ellipse cx="34" cy="92" rx="5" ry="3" fill="#121212" />
        <path d="M60 76 L 64 92" stroke="#121212" stroke-width="4" stroke-linecap="round" />
        <ellipse cx="66" cy="92" rx="5" ry="3" fill="#121212" />

        <!-- Icosaèdre D20 -->
        <polygon points="50,14 80,32 80,68 50,86 20,68 20,32" fill="#FFD300" stroke="#121212" stroke-width="4" />
        <line x1="50" y1="14" x2="50" y2="86" stroke="#121212" stroke-width="2.5" />
        <line x1="20" y1="32" x2="80" y2="68" stroke="#121212" stroke-width="2.5" />
        <line x1="20" y1="68" x2="80" y2="32" stroke="#121212" stroke-width="2.5" />

        <!-- Nombre '20' ou Visage -->
        <rect x="36" y="40" width="28" height="20" rx="6" fill="#FFFFFF" stroke="#121212" stroke-width="2.5" />
        <text x="50" y="55" font-family="Archivo Black" font-size="12" font-weight="900" text-anchor="middle" fill="#121212">20</text>
      </g>

      <!-- ================= 6. MEEPLE (Le Pion Arbitre) ================= -->
      <g v-else>
        <!-- Forme classique Meeple bois -->
        <path
          d="M 50,18
             C 56,18 60,22 60,28
             C 60,33 57,36 54,38
             L 66,46
             L 76,46
             L 76,56
             L 66,56
             L 68,84
             L 54,84
             L 52,66
             L 48,66
             L 46,84
             L 32,84
             L 34,56
             L 24,56
             L 24,46
             L 34,46
             L 46,38
             C 43,36 40,33 40,28
             C 40,22 44,18 50,18 Z"
          fill="#1D4ED8"
          stroke="#121212"
          stroke-width="4"
          stroke-linejoin="round"
        />

        <!-- Casquette ou Sifflet d'arbitre -->
        <path d="M44 18 L 60 14 L 66 18 Z" fill="#FFD300" stroke="#121212" stroke-width="2.5" />
        <!-- Visage Meeple -->
        <circle cx="47" cy="27" r="2.5" fill="#FFFFFF" />
        <circle cx="53" cy="27" r="2.5" fill="#FFFFFF" />
        <circle cx="47" cy="27" r="1" fill="#121212" />
        <circle cx="53" cy="27" r="1" fill="#121212" />
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

export type MascotName = 'dice' | 'domino' | 'card' | 'knight' | 'd20' | 'meeple'
export type MascotMood = 'idle' | 'running' | 'happy' | 'panic' | 'sleeping' | 'dead'
export type MascotSize = 'sm' | 'md' | 'lg' | 'xl'

const props = withDefaults(
  defineProps<{
    name?: MascotName
    mood?: MascotMood
    size?: MascotSize
  }>(),
  {
    name: 'dice',
    mood: 'idle',
    size: 'md',
  }
)

const sizeClass = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'w-8 h-8'
    case 'lg':
      return 'w-20 h-20'
    case 'xl':
      return 'w-32 h-32 sm:w-36 sm:h-36'
    case 'md':
    default:
      return 'w-12 h-12'
  }
})

const animationClass = computed(() => {
  switch (props.mood) {
    case 'running':
    case 'panic':
      return 'animate-mascot-running'
    case 'happy':
      return 'animate-mascot-happy'
    case 'sleeping':
      return 'animate-mascot-sleeping'
    case 'dead':
      return 'opacity-80 rotate-12'
    case 'idle':
    default:
      return ''
  }
})
</script>
