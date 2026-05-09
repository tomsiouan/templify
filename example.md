---
title: Rapport d'exemple
author: Alice Dupont
date: 2026-05-09
header:
  left: "Acme Corp"
  right: "CONFIDENTIEL"
footer:
  right: "{page} / {pages}"
---

## Remerciements

Nous remercions l'ensemble des équipes ayant contribué à la réalisation de ce document, ainsi que nos partenaires pour leur soutien tout au long du projet.

## Introduction

Ce document est un exemple de rapport généré avec **templify**. Il illustre l'ensemble des fonctionnalités disponibles : titres numérotés, tableaux, listes, blocs de code, citations, images avec légende, et bien plus encore.

L'objectif est de fournir un fichier Markdown de référence permettant de valider le rendu visuel du système de templates.

## Contexte et objectifs

### Contexte du projet

Le projet est né d'un besoin de produire des documents PDF de qualité professionnelle à partir de sources Markdown. Les outils existants manquaient soit de flexibilité typographique, soit de contrôle fin sur la mise en page.

### Objectifs

Les objectifs principaux sont les suivants :

1. Générer des PDFs conformes aux standards d'un rapport professionnel
2. Permettre une personnalisation complète via un fichier de configuration
3. Conserver la simplicité d'écriture en Markdown

## Architecture technique

### Vue d'ensemble

Le pipeline de traitement se décompose en trois étapes successives :

1. **Parsing** — lecture du fichier Markdown et extraction du front matter YAML
2. **Templating** — injection du document dans un template HTML
3. **Rendu** — conversion HTML → PDF via Chromium headless (paged.js)

### Composants principaux

| Composant | Package | Rôle |
|---|---|---|
| Parser | `pkg/parser` | Parsing Markdown + front matter |
| Template | `pkg/tmpl` | Chargement et exécution du template |
| Renderer | `pkg/renderer` | Conversion HTML vers PDF |
| Config | `pkg/config` | Chargement et application de la configuration |

### Configuration

La configuration est déclarée dans un fichier YAML. Exemple minimal :

```yaml
font:
  family: Inter
  size: 11pt
  line_height: 1.6

heading_numbers:
  enabled: true
  exclude:
    - Introduction
    - Conclusion

toc:
  enabled: true
  max_depth: 3
```

### Gestion des erreurs

Chaque étape remonte les erreurs avec contexte grâce à `fmt.Errorf("étape: %w", err)`, ce qui permet une traçabilité complète dans les logs.

> Les erreurs critiques interrompent immédiatement le pipeline. Aucun fichier partiel n'est écrit sur le disque.

## Fonctionnalités avancées

### Images avec légende

Les images dont l'attribut `title` est renseigné sont automatiquement encadrées dans un élément `<figure>` avec `<figcaption>` :

![Diagramme de séquence](Markdown-mark.svg "Figure 1 — Flux de traitement")

### Listes de définitions

templify
:   Outil de conversion Markdown → PDF basé sur un système de templates HTML.

paged.js
:   Bibliothèque JavaScript qui implémente le standard CSS Paged Media dans un navigateur.

goldmark
:   Parser Markdown extensible pour Go, conforme à la spécification CommonMark.

### Notes de bas de page

Le format Markdown supporte les notes de bas de page[^1] qui sont automatiquement numérotées et ancrées.
Le format Markdown supporte les notes de bas de page[^2] qui sont automatiquement numérotées et ancrées.

[^1]: Les notes de bas de page apparaissent en bas de la page concernée dans le PDF final.
[^2]: Les notes de bas de page apparaissent en bas de la page concernée dans le PDF final.

### Tableaux avancés

| Fonctionnalité | Statut | Remarque |
|---|---|---|
| Titres numérotés | Disponible | Commence au niveau h2 |
| Sommaire | Disponible | Profondeur configurable |
| Page de couverture | Disponible | Template personnalisable |
| En-tête / pied de page | Disponible | Dégradé CSS supporté |
| Indentation des paragraphes | Disponible | Via `paragraph_indent` |
| Indentation des titres | Disponible | Via `heading_indent` |
| Justification | Disponible | Via `justify: true` |

## Conclusion

Ce fichier d'exemple couvre l'essentiel des constructions Markdown supportées par templify. Il peut servir de base de test pour valider un nouveau template ou une nouvelle configuration.

## Table des illustrations

- Figure 1 — Flux de traitement *(p. 3)*

## Bibliographie

- CommonMark Spec — [spec.commonmark.org](https://spec.commonmark.org)
- CSS Paged Media Module Level 3 — W3C Working Draft

## Sitographie

- Documentation goldmark : [github.com/yuin/goldmark](https://github.com/yuin/goldmark)
- paged.js : [pagedjs.org](https://pagedjs.org)
- go-rod : [go-rod.github.io](https://go-rod.github.io)
