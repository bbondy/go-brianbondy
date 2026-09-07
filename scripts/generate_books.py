import csv
import re
import urllib.parse
import json

def clean_title(title):
    # Remove series information in parentheses
    return re.sub(r'\s*\([^)]*\)', '', title).strip()

def get_book_url(book_id):
    if book_id:
        return f"https://www.goodreads.com/book/show/{book_id}"
    return None

def get_audible_url(title, author):
    """Generate an Audible search URL for the book"""
    search_query = f"{title} {author}".strip()
    encoded_query = urllib.parse.quote(search_query)
    return f"https://www.audible.com/search?keywords={encoded_query}"

def generate_books_manifest():
    books = []
    with open('data/goodreads_library_export.csv', 'r', encoding='utf-8') as file:
        reader = csv.DictReader(file)
        for row in reader:
            title = clean_title(row['Title'])
            authors = []
            if row['Author']:
                authors.append(row['Author'])
            if row['Additional Authors']:
                add_authors = [a.strip() for a in row['Additional Authors'].split(',') if a.strip()]
                authors.extend(add_authors)
            author_text = ", ".join(authors) if authors else "Unknown Author"
            year_published = row['Original Publication Year'] or row['Year Published']
            publisher = row['Publisher']
            pages = row['Number of Pages']
            book_url = get_book_url(row['Book Id'])
            audible_url = get_audible_url(title, author_text)
            book_obj = {
                "title": title,
                "author": author_text,
                "year": int(year_published) if year_published and year_published.isdigit() else None,
                "publisher": publisher,
                "pages": int(pages) if pages and pages.isdigit() else None,
                "goodreads_url": book_url,
                "audible_url": audible_url
            }
            books.append(book_obj)
    with open('data/booksManifest.json', 'w', encoding='utf-8') as f:
        json.dump(books, f, indent=2, ensure_ascii=False)

if __name__ == "__main__":
    generate_books_manifest() 
