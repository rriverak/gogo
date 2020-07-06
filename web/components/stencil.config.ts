import { Config } from '@stencil/core';
import { postcss } from '@stencil/postcss';
import autoprefixer from 'autoprefixer'
import tailwindcss from 'tailwindcss';

export const config: Config = {
  namespace: 'gogo',
  taskQueue: 'async',
  globalStyle: 'src/global/styles.css',  
  plugins: [
    postcss({
      plugins: [
        tailwindcss(),
        autoprefixer(),
      ]
    })
  ] ,
  outputTargets: [
    {
      type: 'dist',
      esmLoaderPath: '../loader'
    },
    {
      type: 'docs-readme'
    },
    {
      type: 'www',
      serviceWorker: null // disable service workers
    }
  ]
};
