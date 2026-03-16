import { ref } from 'vue';
import { ListSubfolders, GetFolderInfo } from '../../wailsjs/go/main/App';
import { fs as models } from '../../wailsjs/go/models';
import { LogError } from '../../wailsjs/runtime/runtime';

export function useFolderTree() {
  const folderTreeRoot = ref<models.Folder | null>(null);
  const isTreeLoading = ref<boolean>(false);
  const treeError = ref<string>("");

  async function loadFolderTree(basePath: string) {
    if (!basePath) return;
    isTreeLoading.value = true;
    treeError.value = "";
    folderTreeRoot.value = null;
    try {
      // Get root info
      const root = await GetFolderInfo(basePath);
      // Load immediate children for the root
      root.children = await ListSubfolders(basePath);
      folderTreeRoot.value = root;
    } catch (err: any) {
      const errMsg = `Failed to load folder tree: ${err.message || err}`;
      LogError(errMsg);
      treeError.value = errMsg;
    } finally {
      isTreeLoading.value = false;
    }
  }

  async function loadSubfolders(folder: models.Folder) {
    if (folder.children && folder.children.length > 0) return; // Already loaded
    if (!folder.hasChildren) return;

    try {
      const children = await ListSubfolders(folder.path);
      folder.children = children;
    } catch (err: any) {
      LogError(`Failed to load subfolders for ${folder.path}: ${err}`);
    }
  }

  function clearTree() {
    folderTreeRoot.value = null;
    treeError.value = "";
  }

  return {
    folderTreeRoot,
    isTreeLoading,
    treeError,
    loadFolderTree,
    loadSubfolders,
    clearTree,
  };
}
