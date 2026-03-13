import { performance } from 'perf_hooks';

// Simulate a large directory structure
const TOTAL_FOLDERS = 10000;
const LEAF_INTERVAL = 10; // Every 10th folder is a leaf

const flatFolderList = [];
const leafFolderList = [];

for (let i = 0; i < TOTAL_FOLDERS; i++) {
  const folderName = `folder_${i}`;
  flatFolderList.push(folderName);
  if (i % LEAF_INTERVAL === 0) {
    leafFolderList.push(folderName);
  }
}

// Current folder is near the beginning
const currentFolderIndex = 1;

// --- Baseline (Array includes) ---
function findNextLeafArrayIncludes() {
  let found = false;
  for (let i = currentFolderIndex + 1; i < flatFolderList.length; i++) {
    const potentialNextFolder = flatFolderList[i];
    if (leafFolderList.includes(potentialNextFolder)) {
      found = true;
      break;
    }
  }
  return found;
}

// --- Optimized (Set has) ---
// Pre-calculate Set (simulating a computed property)
const leafFolderSet = new Set(leafFolderList);

function findNextLeafSetHas() {
  let found = false;
  for (let i = currentFolderIndex + 1; i < flatFolderList.length; i++) {
    const potentialNextFolder = flatFolderList[i];
    if (leafFolderSet.has(potentialNextFolder)) {
      found = true;
      break;
    }
  }
  return found;
}

// --- Benchmark ---
const ITERATIONS = 10000; // Number of times to run the check to get measurable time

console.log(`Benchmarking with ${TOTAL_FOLDERS} folders, ${leafFolderList.length} leaf folders.`);
console.log(`Running ${ITERATIONS} iterations...\n`);

// Baseline
const startArray = performance.now();
for (let i = 0; i < ITERATIONS; i++) {
  findNextLeafArrayIncludes();
}
const endArray = performance.now();
const timeArray = endArray - startArray;
console.log(`Baseline (Array.includes): ${timeArray.toFixed(2)} ms`);

// Optimized
const startSet = performance.now();
for (let i = 0; i < ITERATIONS; i++) {
  findNextLeafSetHas();
}
const endSet = performance.now();
const timeSet = endSet - startSet;
console.log(`Optimized (Set.has):      ${timeSet.toFixed(2)} ms`);

// Improvement
const improvement = ((timeArray - timeSet) / timeArray) * 100;
const speedup = timeArray / timeSet;
console.log(`\nImprovement: ${improvement.toFixed(2)}% faster (${speedup.toFixed(2)}x)`);
