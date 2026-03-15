import { test, expect } from '@playwright/test';

test('should have Dracula theme background color', async ({ page }) => {
  await page.goto('/');
  
  // Dracula background: #282a36 (rgb(40, 42, 54))
  const backgroundColor = await page.evaluate(() => {
    return window.getComputedStyle(document.body).backgroundColor;
  });
  
  expect(backgroundColor).toBe('rgb(40, 42, 54)');
});

test('sidebar should have Dracula background color', async ({ page }) => {
  await page.goto('/');
  
  const sidebar = page.locator('.sidebar');
  const backgroundColor = await sidebar.evaluate((el) => {
    return window.getComputedStyle(el).backgroundColor;
  });
  
  expect(backgroundColor).toBe('rgb(40, 42, 54)');
});
