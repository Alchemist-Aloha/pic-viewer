import { ref, computed } from 'vue';
import { ListSubfolders } from '../../wailsjs/go/main/App';
import { fs as models } from '../../wailsjs/go/models';
import { LogError } from '../../wailsjs/runtime/runtime';

export function useFolderTree() {
  const folderTreeRoot = ref<models.Folder | null>(null);
  const flatFolderList = ref<string[]>([]);
  const leafFolderList = ref<string[]>([]);
  const isTreeLoading = ref<boolean>(false);
  const treeError = ref<string>("");

  const leafFolderSet = computed(() => new Set(leafFolderList.value));

  const flatFolderIndexMap = computed(() => {
    const map = new Map<string, number>();
    flatFolderList.value.forEach((path, index) => map.set(path, index));
    return map;
  });

  const leafFolderIndexMap = computed(() => {
    const map = new Map<string, number>();
    leafFolderList.value.forEach((path, index) => map.set(path, index));
    return map;
  });

  function flattenTree(node: models.Folder | null): string[] {
    if (!node) return [];
    let paths: string[] = [node.path];
    if (node.children) {
      for (const child of node.children) {
        paths = paths.concat(flattenTree(child));
      }
    }
    return paths;
  }

  function getLeafFolders(node: models.Folder | null): string[] {
    if (!node) return [];
    let leafPaths: string[] = [];
    if (!node.children || node.children.length === 0) {
      leafPaths.push(node.path);
    } else {
      for (const child of node.children) {
        leafPaths = leafPaths.concat(getLeafFolders(child));
      }
    }
    return leafPaths;
  }

  async function loadFolderTree(basePath: string) {
    if (!basePath) return;
    isTreeLoading.value = true;
    treeError.value = "";
    folderTreeRoot.value = null;
    flatFolderList.value = [];
    leafFolderList.value = [];
    try {
      const tree = await ListSubfolders(basePath);
      folderTreeRoot.value = tree;
      flatFolderList.value = flattenTree(tree);
      leafFolderList.value = getLeafFolders(tree);
    } catch (err: any) {
      const errMsg = `Failed to load folder tree: ${err.message || err}`;
      LogError(errMsg);
      treeError.value = errMsg;
    } finally {
      isTreeLoading.value = false;
    }
  }

  function clearTree() {
    folderTreeRoot.value = null;
    flatFolderList.value = [];
    leafFolderList.value = [];
    treeError.value = "";
  }

  return {
    folderTreeRoot,
    flatFolderList,
    leafFolderList,
    leafFolderSet,
    flatFolderIndexMap,
    leafFolderIndexMap,
    isTreeLoading,
    treeError,
    loadFolderTree,
    clearTree,
  };
}
