import csv
import re
import urllib.parse

def clean_title(title):
    # Remove series information in parentheses
    return re.sub(r'\s*\([^)]*\)', '', title).strip()

def get_book_url(book_id):
    if book_id:
        return f"https://www.goodreads.com/book/show/{book_id}"
    return None

def get_audible_url(title, author):
    """Generate an Audible search URL for the book"""
    # Clean the title and author for search
    search_query = f"{title} {author}".strip()
    # URL encode the search query
    encoded_query = urllib.parse.quote(search_query)
    return f"https://www.audible.com/search?keywords={encoded_query}"

def generate_books_markdown():
    books = []  # List to store all books
    
    # Read the CSV file
    with open('data/goodreads_library_export.csv', 'r', encoding='utf-8') as file:
        reader = csv.DictReader(file)
        for row in reader:
            # Clean up the title and author
            title = clean_title(row['Title'])
            authors = []
            if row['Author']:
                authors.append(row['Author'])
            if row['Additional Authors']:
                # Split and clean additional authors
                add_authors = [a.strip() for a in row['Additional Authors'].split(',') if a.strip()]
                authors.extend(add_authors)
            
            author_text = ", ".join(authors) if authors else "Unknown Author"
            
            # Get publication info
            year_published = row['Original Publication Year'] or row['Year Published']
            publisher = row['Publisher']
            pages = row['Number of Pages']
            
            # Get book URLs
            book_url = get_book_url(row['Book Id'])
            audible_url = get_audible_url(title, author_text)
            
            # Create book entry with links using project-style structure
            book_entry = f"""<div class="card-entry">
  <div class="project-entry">
    <div class="project-content">
      <h3>
        <a href="{book_url}" target="_blank">{title}</a>
      </h3>
      <div class="card-details">
        <div class="book-author">by {author_text}</div>
        <div class="book-meta">{year_published}{f", {publisher}" if publisher else ""}{f", {pages} pages" if pages else ""}</div>
      </div>
      <div class="project-links">
        <a href="{book_url}" target="_blank" title="View on Goodreads" class="action-link website-link"><i class="fas fa-book"></i></a>
        <a href="{audible_url}" target="_blank" title="Search on Audible" class="action-link audible-link"><i class="fab fa-audible"></i></a>
      </div>
    </div>
  </div>
</div>"""
            books.append(book_entry)
    
    # Generate markdown content
    markdown_content = """# Books

The list below contains books that I've read and good enough to share.

"""
    
    # Add books in original order
    for book in books:
        markdown_content += book + "\n"
    
    # Write to file
    with open('data/markdown/books.markdown', 'w', encoding='utf-8') as file:
        file.write(markdown_content)

if __name__ == "__main__":
    generate_books_markdown() 
